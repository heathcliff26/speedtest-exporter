package collector

import (
	"log/slog"
	"sync"

	"github.com/heathcliff26/speedtest-exporter/pkg/cache"
	"github.com/heathcliff26/speedtest-exporter/pkg/speedtest"
	"github.com/prometheus/client_golang/prometheus"
)

type Collector struct {
	cache     *cache.Cache
	speedtest speedtest.Speedtest
	instance  string
}

var (
	variableLabels    = []string{"ip", "isp", "instance"}
	jitterLatencyDesc = prometheus.NewDesc("speedtest_jitter_latency_milliseconds", "Speedtest current Jitter in ms", variableLabels, nil)
	pingDesc          = prometheus.NewDesc("speedtest_ping_latency_milliseconds", "Speedtest current Ping in ms", variableLabels, nil)
	downloadSpeedDesc = prometheus.NewDesc("speedtest_download_megabits_per_second", "Speedtest current Download Speed in Mbit/s", variableLabels, nil)
	uploadSpeedDesc   = prometheus.NewDesc("speedtest_upload_megabits_per_second", "Speedtest current Upload Speed in Mbit/s", variableLabels, nil)
	dataUsedDesc      = prometheus.NewDesc("speedtest_data_used_megabytes", "Data used for speedtest in MB", variableLabels, nil)
	durationDesc      = prometheus.NewDesc("speedtest_duration_milliseconds", "Duration of the speedtest in milliseconds", variableLabels, nil)
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
func NewCollector(cache *cache.Cache, speedtest speedtest.Speedtest, instance string) (*Collector, error) {
	if speedtest == nil {
		return nil, ErrNoSpeedtest{}
	}
	return &Collector{
		cache:     cache,
		speedtest: speedtest,
		instance:  instance,
	}, nil
}

// Implements the Describe function for prometheus.Collector
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Concurrency safe function to get the latest result of the speedtest.
// Will either return the cached result or run a new test.
func (c *Collector) getSpeedtestResult() *speedtest.SpeedtestResult {
	// Lock here to prevent running more than one Speedtest at a time, since they would affect each others results
	speedtestMutex.Lock()
	defer speedtestMutex.Unlock()

	result, ok := c.cache.Read()
	if ok {
		slog.Debug("Cache has not expired, returning cached results", slog.String("expires", c.cache.ExpiresAt().Local().String()))
		return result
	}
	slog.Debug("Cache expired, running new Speedtest")
	result = c.speedtest.Speedtest()
	c.cache.Save(result)
	return result
}

// Implements the Collect function for prometheus.Collector
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	slog.Debug("Starting collection of speedtest metrics")
	result := c.getSpeedtestResult()
	var up float64
	if result.Success() {
		up = 1
		labelValues := []string{result.ClientIP(), result.ClientISP(), c.instance}
		ch <- prometheus.MustNewConstMetric(jitterLatencyDesc, prometheus.GaugeValue, result.JitterLatency(), labelValues...)
		ch <- prometheus.MustNewConstMetric(pingDesc, prometheus.GaugeValue, result.Ping(), labelValues...)
		ch <- prometheus.MustNewConstMetric(downloadSpeedDesc, prometheus.GaugeValue, result.DownloadSpeed(), labelValues...)
		ch <- prometheus.MustNewConstMetric(uploadSpeedDesc, prometheus.GaugeValue, result.UploadSpeed(), labelValues...)
		ch <- prometheus.MustNewConstMetric(dataUsedDesc, prometheus.GaugeValue, result.DataUsed(), labelValues...)
		ch <- prometheus.MustNewConstMetric(durationDesc, prometheus.GaugeValue, float64(result.Duration()), labelValues...)
	}
	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, up)
	slog.Debug("Finished collection of speedtest metrics")
}
