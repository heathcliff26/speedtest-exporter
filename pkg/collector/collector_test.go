package collector

import (
	"testing"
	"time"

	"github.com/heathcliff26/speedtest-exporter/pkg/cache"
	"github.com/heathcliff26/speedtest-exporter/pkg/speedtest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mockSpeedtestResult = speedtest.NewSpeedtestResult(0.5, 15, 876.53, 12.34, 950.3079, "1234", "example.org", "Foo Corp.", "127.0.0.1")

const defaultCacheTime = 5 * time.Minute

func NewMockSpeedtest() *speedtest.MockSpeedtest {
	return &speedtest.MockSpeedtest{Result: mockSpeedtestResult}
}

func TestNewCollector(t *testing.T) {
	s := NewMockSpeedtest()
	c := cache.NewCache(false, "", defaultCacheTime)
	expectedCollector := &Collector{
		cache:     c,
		speedtest: s,
		instance:  "testinstance",
	}

	actualCollector, err := NewCollector(c, s, "testinstance")
	require.NoError(t, err, "Should create new Collector")

	assert := assert.New(t)

	assert.Equal(expectedCollector, actualCollector)

	_, err = NewCollector(nil, nil, "testinstance")
	assert.Equal(ErrNoSpeedtest{}, err)
}

func TestResultFromCache(t *testing.T) {
	speedtestRan := false
	s := NewMockSpeedtest()
	s.Callback = func() {
		speedtestRan = true
	}

	c, err := NewCollector(cache.NewCache(false, "", defaultCacheTime), s, "testinstance")
	require.NoError(t, err, "Should create new Collector")

	c.cache.Save(speedtest.NewFailedSpeedtestResult())

	result := c.getSpeedtestResult()

	assert := assert.New(t)

	assert.Equal(speedtest.NewFailedSpeedtestResult(), result)
	assert.False(speedtestRan, "Should not have called the mock speedtest")
}

func TestRunSpeedtestWhenCacheExpired(t *testing.T) {
	speedtestRan := false
	s := NewMockSpeedtest()
	s.Callback = func() {
		speedtestRan = true
	}

	c, err := NewCollector(cache.NewCache(false, "", defaultCacheTime), s, "testinstance")
	require.NoError(t, err, "Should create new Collector")
	result := c.getSpeedtestResult()

	assert := assert.New(t)

	assert.Equal(mockSpeedtestResult, result, "Should return the mock speedtest result")
	cachedResult, valid := c.cache.Read()
	assert.True(valid, "Cache should be valid after speedtest run")
	assert.Equal(cachedResult, result, "Cached result should equal returned result")

	assert.True(speedtestRan, "Should have called the mock speedtest")
}

func TestSpeedtestIsNotRunConcurrently(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	i := 0
	sleeping := make(chan bool, 1)
	s := NewMockSpeedtest()
	s.Callback = func() {
		i++
		sleeping <- true
		time.Sleep(10 * time.Second)
	}

	c, err := NewCollector(cache.NewCache(false, "", defaultCacheTime), s, "testinstance")
	require.NoError(t, err, "Should create new Collector")

	assert := assert.New(t)

	var result1, result2 *speedtest.SpeedtestResult
	go func() {
		result1 = c.getSpeedtestResult()
	}()
	<-sleeping
	cachedResult, _ := c.cache.Read()
	assert.Nil(cachedResult)
	result2 = c.getSpeedtestResult()

	cachedResult, _ = c.cache.Read()

	assert.NotNil(result1)
	assert.Equal(result1, result2)
	assert.Equal(result1, cachedResult)
	assert.Equal(1, i)
}

func TestCollect(t *testing.T) {
	s := NewMockSpeedtest()
	c, err := NewCollector(nil, s, "testinstance")
	require.NoError(t, err, "Should create new Collector")

	t.Run("Success", func(t *testing.T) {
		ch := make(chan prometheus.Metric, 1)
		go c.Collect(ch)

		actualLabelValues := []string{mockSpeedtestResult.ClientIP(), mockSpeedtestResult.ClientISP(), "testinstance"}

		actualMetric := <-ch
		expectedMetric := prometheus.MustNewConstMetric(jitterLatencyDesc, prometheus.GaugeValue, mockSpeedtestResult.JitterLatency(), actualLabelValues...)
		assert.Equal(t, expectedMetric, actualMetric)

		actualMetric = <-ch
		expectedMetric = prometheus.MustNewConstMetric(pingDesc, prometheus.GaugeValue, mockSpeedtestResult.Ping(), actualLabelValues...)
		assert.Equal(t, expectedMetric, actualMetric)

		actualMetric = <-ch
		expectedMetric = prometheus.MustNewConstMetric(downloadSpeedDesc, prometheus.GaugeValue, mockSpeedtestResult.DownloadSpeed(), actualLabelValues...)
		assert.Equal(t, expectedMetric, actualMetric)

		actualMetric = <-ch
		expectedMetric = prometheus.MustNewConstMetric(uploadSpeedDesc, prometheus.GaugeValue, mockSpeedtestResult.UploadSpeed(), actualLabelValues...)
		assert.Equal(t, expectedMetric, actualMetric)

		actualMetric = <-ch
		expectedMetric = prometheus.MustNewConstMetric(dataUsedDesc, prometheus.GaugeValue, mockSpeedtestResult.DataUsed(), actualLabelValues...)
		assert.Equal(t, expectedMetric, actualMetric)

		actualMetric = <-ch
		expectedMetric = prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 1)
		assert.Equal(t, expectedMetric, actualMetric)
	})

	t.Run("Failure", func(t *testing.T) {
		ch := make(chan prometheus.Metric, 1)
		s.Fail = true
		go c.Collect(ch)
		actualMetric := <-ch
		expectedMetric := prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, 0)
		assert.Equal(t, expectedMetric, actualMetric)
	})
}
