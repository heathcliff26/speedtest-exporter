package promremote

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prometheus/model/timestamp"
	"github.com/prometheus/prometheus/prompb"
)

type Client struct {
	endpoint string
	instance string
	job      string
	username string
	password string
	registry *prometheus.Registry
}

type MetricData struct {
	Labels    map[string]string
	Timestamp time.Time
	Value     float64
}

func NewWriteClient(endpoint, instance, job string, reg *prometheus.Registry) (*Client, error) {
	if endpoint == "" {
		return nil, ErrMissingEndpoint{}
	}
	if instance == "" {
		return nil, ErrMissingInstance{}
	}
	if job == "" {
		return nil, ErrMissingJob{}
	}
	if reg == nil {
		return nil, ErrMissingRegistry{}
	}
	return &Client{
		endpoint: endpoint,
		instance: instance,
		job:      job,
		registry: reg,
	}, nil
}

func (c *Client) Endpoint() string {
	if c == nil {
		return ""
	}
	return c.endpoint
}

func (c *Client) Registry() *prometheus.Registry {
	if c == nil {
		return nil
	}
	return c.registry
}

// Set credentials needed for basic auth, return error if not provided
func (c *Client) SetBasicAuth(username, password string) error {
	if username == "" || password == "" {
		return ErrMissingAuthCredentials{}
	}
	c.username = username
	c.password = password
	return nil
}

// Send TimeSeries to remote_write endpoint
func (c *Client) post(ts []prompb.TimeSeries) error {
	wr := prompb.WriteRequest{Timeseries: ts}
	data, err := wr.Marshal()
	if err != nil {
		return err
	}
	body := snappy.Encode(nil, data)

	req, err := http.NewRequest(http.MethodPost, c.Endpoint(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Encoding", "snappy")
	req.Header.Add("Content-Type", "application/x-protobuf")
	req.Header.Set("X-Prometheus-Remote-Read-Version", "0.1.0")
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	httpClient := http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return NewErrRemoteWriteFailed(res.StatusCode, req.Body)
	}

	return nil
}

// Collect metrics from registry and convert them to TimeSeries
func (c *Client) collect() ([]prompb.TimeSeries, error) {
	ch := make(chan prometheus.Metric)
	go func() {
		c.registry.Collect(ch)
		close(ch)
	}()

	var res []prompb.TimeSeries
	for metric := range ch {
		// Extract name of metric
		regex := regexp.MustCompile("fqName: \"([a-zA-Z_:][a-zA-Z0-9_:]*)\"")
		fqName := regex.FindStringSubmatch(metric.Desc().String())
		if len(fqName) < 2 {
			return nil, &ErrInvalidMetricDesc{Desc: metric.Desc().String()}
		}

		// Convert metric to readable format
		m := &dto.Metric{}
		err := metric.Write(m)
		if err != nil {
			return nil, err
		}

		// Extract labels
		labels := make([]prompb.Label, 0, len(m.Label)+3)
		labels = append(labels, prompb.Label{
			Name:  "__name__",
			Value: fqName[1],
		})
		labels = append(labels, prompb.Label{
			Name:  "instance",
			Value: c.instance,
		})
		labels = append(labels, prompb.Label{
			Name:  "job",
			Value: c.job,
		})
		dropLabels := []string{"__name__", "instance", "job"}
		for _, l := range m.Label {
			if !slices.Contains(dropLabels, l.GetName()) {
				labels = append(labels, prompb.Label{
					Name:  l.GetName(),
					Value: l.GetValue(),
				})
			}
		}

		ts := prompb.TimeSeries{
			Labels: labels,
		}

		// Extract value and timestamp
		var value float64
		if m.Counter != nil {
			value = m.Counter.GetValue()
		} else if m.Gauge != nil {
			value = m.Gauge.GetValue()
		} else if m.Untyped != nil {
			value = m.Untyped.GetValue()
		} else {
			return nil, fmt.Errorf("unknown metric type")
		}
		ts.Samples = []prompb.Sample{
			{
				Value:     value,
				Timestamp: timestamp.FromTime(time.Now()),
			},
		}

		res = append(res, ts)
	}
	return res, nil
}

// Collect metrics and send them to remote server in interval.
// Does not block main thread execution
func (c *Client) Run(interval time.Duration, quit chan bool) {
	go func() {
		for {
			ts, err := c.collect()
			if err != nil {
				slog.Error("Failed to collect metrics for remote_write", "err", err)
			}
			err = c.post(ts)
			if err != nil {
				slog.Error("Failed to send metrics to remote endpoint", "err", err)
			} else {
				slog.Debug("Successfully send metrics via remote_write")
			}

			var elapsedTime time.Duration = 0
			for elapsedTime < interval {
				timer := time.NewTimer(1 * time.Second)
				select {
				case <-timer.C:
					elapsedTime += time.Duration(1 * time.Second)
				case <-quit:
					timer.Stop()
					slog.Info("Received stop signal, shutting down remote_write client")
					return
				}
			}
		}
	}()
}
