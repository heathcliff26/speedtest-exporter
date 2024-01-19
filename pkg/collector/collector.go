package collector

import (
	"log/slog"
	"sync"
	"time"

	"github.com/heathcliff26/speedtest-exporter/pkg/speedtest"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	cacheTime     time.Duration
	speedtest     speedtest.Speedtest
	lastResult    *speedtest.SpeedtestResult
	nextSpeedtest time.Time
}

var (
	variableLabels    = []string{"ip", "isp"}
	jitterLatencyDesc = prometheus.NewDesc("speedtest_jitter_latency_milliseconds", "Speedtest current Jitter in ms", variableLabels, nil)
	pingDesc          = prometheus.NewDesc("speedtest_ping_latency_milliseconds", "Speedtest current Ping in ms", variableLabels, nil)
	downloadSpeedDesc = prometheus.NewDesc("speedtest_download_megabits_per_second", "Speedtest current Download Speed in Mbit/s", variableLabels, nil)
	uploadSpeedDesc   = prometheus.NewDesc("speedtest_upload_megabits_per_second", "Speedtest current Upload Speed in Mbit/s", variableLabels, nil)
	dataUsedDesc      = prometheus.NewDesc("speedtest_data_used_megabytes", "Data used for speedtest in MB", variableLabels, nil)
	upDesc            = prometheus.NewDesc("speedtest_up", "Indicates if the speedtest was successful", nil, nil)
)

// Used to prevent concurrent runs of Speedtest's
var speedtestMutex sync.Mutex

// Create new instance of collector, returns error if an instance of speedtest is not provided
// Arguments:
//
//	cacheTime: Minimum time between speedtest runs
//	instance: Name of this instance, provided as label on all metrics
//	speedtest: Instance of speedtest to use for collection metrics
func NewCollector(cacheTime time.Duration, speedtest speedtest.Speedtest) (*Collector, error) {
	if speedtest == nil {
		return nil, ErrNoSpeedtest{}
	}
	return &Collector{
		cacheTime: cacheTime,
		speedtest: speedtest,
	}, nil
}

// Implements the Describe function for prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Reset time for new cache
func (c *Collector) setNextSpeedtestTime() {
	c.nextSpeedtest = time.Now().Add(time.Minute * c.cacheTime)
	slog.Debug("Next Speedtest will not be executed before", slog.String("time", c.nextSpeedtest.Local().String()))
}

// Concurrency safe function to get the latest result of the speedtest.
// Will either return the cached result or run a new test.
func (c *Collector) getSpeedtestResult() *speedtest.SpeedtestResult {
	if c.lastResult != nil && time.Now().Before(c.nextSpeedtest) {
		slog.Debug("Cache has not expired, returning cached results")
		return c.lastResult
	}

	// Lock here to prevent running more than one Speedtest at a time, since they would affect each others results
	speedtestMutex.Lock()
	defer speedtestMutex.Unlock()
	// Check again if another thread already ran a Speedtest while this one waited
	if c.lastResult != nil && time.Now().Before(c.nextSpeedtest) {
		slog.Debug("Cache has been renewed, returning cached results")
		return c.lastResult
	}
	slog.Debug("Cache expired, running new Speedtest")
	c.lastResult = c.speedtest.Speedtest()
	c.setNextSpeedtestTime()
	return c.lastResult
}

// Implements the Collect function for prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	slog.Debug("Starting collection of speedtest metrics")
	result := c.getSpeedtestResult()
	var up float64
	if result.Success() {
		up = 1
		labelValues := []string{result.ClientIp(), result.ClientIsp()}
		ch <- prometheus.MustNewConstMetric(jitterLatencyDesc, prometheus.GaugeValue, result.JitterLatency(), labelValues...)
		ch <- prometheus.MustNewConstMetric(pingDesc, prometheus.GaugeValue, result.Ping(), labelValues...)
		ch <- prometheus.MustNewConstMetric(downloadSpeedDesc, prometheus.GaugeValue, result.DownloadSpeed(), labelValues...)
		ch <- prometheus.MustNewConstMetric(uploadSpeedDesc, prometheus.GaugeValue, result.UploadSpeed(), labelValues...)
		ch <- prometheus.MustNewConstMetric(dataUsedDesc, prometheus.GaugeValue, result.DataUsed(), labelValues...)
	}
	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, up)
	slog.Debug("Finished collection of speedtest metrics")
}
