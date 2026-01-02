package promremote

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/exp/api/remote"
	writev2 "github.com/prometheus/client_golang/exp/api/remote/genproto/v2"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var (
	fqNameRegex = regexp.MustCompile("fqName: \"([a-zA-Z_:][a-zA-Z0-9_:]*)\"")
	helpRegex   = regexp.MustCompile("help: \"([^\"]*)\"")
)

const httpClientTimeout = 10 * time.Second
const defaultJobName = "promremote"

type Client interface {
	// Registry returns the Prometheus registry used by the client.
	Registry() *prometheus.Registry
	// Run periodically collects metrics and sends them to the remote server.
	// It runs as a background goroutine and does not block the calling thread.
	Run(interval time.Duration) error
	// IsRunning returns true if the client is currently running.
	IsRunning() bool
	// Stop stops the remote_write client.
	Stop()
}

type client struct {
	instance string
	job      string
	client   *http.Client
	registry *prometheus.Registry

	remoteAPI *remote.API
	running   atomic.Bool
	cancel    context.CancelFunc
}

type ClientOption func(*client) error

// WithBasicAuth configures basic authentication when sending metrics.
// Returns an error if username or password are empty.
func WithBasicAuth(username, password string) ClientOption {
	return func(c *client) error {
		if username == "" || password == "" {
			return ErrMissingAuthCredentials{}
		}
		addBasicAuthToHTTPClient(c.client, username, password)
		return nil
	}
}

// WithInstanceLabel sets the instance label for the metrics.
// By default the hostname of the machine is used.
func WithInstanceLabel(instance string) ClientOption {
	return func(c *client) error {
		if instance == "" {
			return ErrMissingInstance{}
		}
		c.instance = instance
		return nil
	}
}

// WithJobLabel sets the job label for the metrics.
// By default "promremote" is used.
func WithJobLabel(job string) ClientOption {
	return func(c *client) error {
		if job == "" {
			return ErrMissingJob{}
		}
		c.job = job
		return nil
	}
}

// NewWriteClient creates a new remote_write client.
// Parameters:
//   - endpoint: URL of the remote_write endpoint
//   - reg: Prometheus registry to collect metrics from
//   - opts: optional client options
func NewWriteClient(endpoint string, reg *prometheus.Registry, opts ...ClientOption) (Client, error) {
	if endpoint == "" {
		return nil, ErrMissingEndpoint{}
	}
	if reg == nil {
		return nil, ErrMissingRegistry{}
	}

	c := &client{
		instance: getHostname(),
		job:      defaultJobName,
		client:   newHTTPClientWithTimeout(),
		registry: reg,
	}
	var err error
	for _, opt := range opts {
		err = opt(c)
		if err != nil {
			return nil, err
		}
	}

	// Set an empty path to ensure that our own path is not overridden with the default path.
	// Disable retries by setting MaxRetries to -1.
	c.remoteAPI, err = remote.NewAPI(endpoint, remote.WithAPIPath(""), remote.WithAPIHTTPClient(c.client), remote.WithAPIBackoff(remote.BackoffConfig{MaxRetries: -1}))
	if err != nil {
		return nil, NewErrFailedToCreateRemoteAPI(err)
	}

	return c, nil
}

// Implement interface method
func (c *client) Registry() *prometheus.Registry {
	if c == nil {
		return nil
	}
	return c.registry
}

// Collect metrics from registry and convert them to TimeSeries
func (c *client) collect() (*writev2.Request, error) {
	ch := make(chan prometheus.Metric)
	go func() {
		c.registry.Collect(ch)
		close(ch)
	}()

	res := &writev2.Request{}
	s := writev2.NewSymbolTable()

	for metric := range ch {
		// Extract name of metric
		fqName := fqNameRegex.FindStringSubmatch(metric.Desc().String())
		if len(fqName) < 2 {
			return nil, &ErrInvalidMetricDesc{Desc: metric.Desc().String()}
		}
		helpRef := helpRegex.FindStringSubmatch(metric.Desc().String())
		help := ""
		if len(helpRef) == 2 {
			help = helpRef[1]
		}

		// Convert metric to readable format
		m := &dto.Metric{}
		err := metric.Write(m)
		if err != nil {
			return nil, err
		}

		// Extract labels
		labels := make([]string, 0, 2*(len(m.Label)+3))
		labels = append(labels, "__name__", fqName[1], "instance", c.instance, "job", c.job)
		dropLabels := []string{"__name__", "instance", "job"}
		for _, l := range m.Label {
			if !slices.Contains(dropLabels, l.GetName()) {
				labels = append(labels, l.GetName(), l.GetValue())
			}
		}

		// Extract value and timestamp
		var metricType writev2.Metadata_MetricType
		var value float64
		if m.Counter != nil {
			metricType = writev2.Metadata_METRIC_TYPE_COUNTER
			value = m.Counter.GetValue()
		} else if m.Gauge != nil {
			metricType = writev2.Metadata_METRIC_TYPE_GAUGE
			value = m.Gauge.GetValue()
		} else if m.Untyped != nil {
			metricType = writev2.Metadata_METRIC_TYPE_UNSPECIFIED
			value = m.Untyped.GetValue()
		} else {
			return nil, fmt.Errorf("unknown metric type")
		}

		ts := &writev2.TimeSeries{
			Metadata: &writev2.Metadata{
				Type:    metricType,
				HelpRef: s.Symbolize(help),
			},
			LabelsRefs: s.SymbolizeLabels(labels, nil),
			Samples: []*writev2.Sample{
				{
					Value:     value,
					Timestamp: time.Now().UnixMilli(),
				},
			},
		}

		res.Timeseries = append(res.Timeseries, ts)
	}
	res.Symbols = s.Symbols()
	return res, nil
}

// Implement interface method
func (c *client) Run(interval time.Duration) error {
	if !c.running.CompareAndSwap(false, true) {
		return ErrClientAlreadyRunning{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer c.running.Store(false)
		slog.Debug("Starting remote_write client")
		for {
			req, err := c.collect()
			if err != nil {
				slog.Error("Failed to collect metrics for remote_write", "err", err)
			}

			stats, err := c.remoteAPI.Write(ctx, remote.WriteV2MessageType, req)
			if err != nil {
				slog.Error("Failed to send metrics to remote endpoint", "err", err)
			} else {
				slog.Debug("Successfully sent metrics via remote_write", slog.Int("count", stats.AllSamples()), slog.Bool("written", !stats.NoDataWritten()))
			}
			select {
			case <-ticker.C:

			case <-ctx.Done():
				slog.Info("Stopping remote_write client")
				return
			}
		}
	}()

	return nil
}

// Implement interface method
func (c *client) IsRunning() bool {
	return c.running.Load()
}

// Implement interface method
func (c *client) Stop() {
	if c.IsRunning() && c.cancel != nil {
		c.cancel()
	}
}
