package collector

import (
	"testing"
	"time"

	"github.com/heathcliff26/speedtest-exporter/pkg/speedtest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

var mockSpeedtestResult = speedtest.NewSpeedtestResult(0.5, 15, 876.53, 12.34, 950.3079, "1234", "example.org", "Foo Corp.", "127.0.0.1")

const defaultCacheTime = 5 * time.Minute

func NewMockSpeedtest() *speedtest.MockSpeedtest {
	return &speedtest.MockSpeedtest{Result: mockSpeedtestResult}
}

func TestNewCollector(t *testing.T) {
	s := NewMockSpeedtest()
	expectedCollector := &Collector{
		cacheTime: defaultCacheTime,
		speedtest: s,
	}

	actualCollector, err := NewCollector(defaultCacheTime, s)
	if err != nil {
		t.Fatalf("Could not create new Collector: %v", err)
	}

	assert := assert.New(t)

	assert.Equal(expectedCollector, actualCollector)

	_, err = NewCollector(defaultCacheTime, nil)
	assert.Equal(ErrNoSpeedtest{}, err)
}

func TestSetNextSpeedtestTime(t *testing.T) {
	now := time.Now()

	c, err := NewCollector(defaultCacheTime, NewMockSpeedtest())
	if err != nil {
		t.Fatalf("Could not create new Collector: %v", err)
	}

	c.nextSpeedtest = now
	c.setNextSpeedtestTime()

	assert.Greater(t, c.nextSpeedtest, now.Add(defaultCacheTime-time.Millisecond))
	assert.Less(t, c.nextSpeedtest, now.Add(defaultCacheTime+time.Millisecond))
}

func TestFirstSpeedtestRun(t *testing.T) {
	c, err := NewCollector(defaultCacheTime, NewMockSpeedtest())
	if err != nil {
		t.Fatalf("Could not create new Collector: %v", err)
	}
	result := c.getSpeedtestResult()

	assert := assert.New(t)

	assert.Equal(mockSpeedtestResult, result)
	assert.Equal(result, c.lastResult)
	assert.NotEmpty(c.nextSpeedtest)
}

func TestResultFromCache(t *testing.T) {
	speedtestRan := false
	s := NewMockSpeedtest()
	s.Callback = func() {
		speedtestRan = true
	}

	c, err := NewCollector(defaultCacheTime, s)
	if err != nil {
		t.Fatalf("Could not create new Collector: %v", err)
	}
	c.lastResult = speedtest.NewFailedSpeedtestResult()
	c.nextSpeedtest = time.Now().Add(time.Hour)

	result := c.getSpeedtestResult()

	assert := assert.New(t)

	assert.NotNil(result)
	assert.Equal(speedtest.NewFailedSpeedtestResult(), result)
	if speedtestRan {
		t.Error("Speedtest has been called")
	}
}

func TestRunSpeedtestWhenCacheEmpty(t *testing.T) {
	speedtestRan := false
	s := NewMockSpeedtest()
	s.Callback = func() {
		speedtestRan = true
	}

	c, err := NewCollector(defaultCacheTime, s)
	if err != nil {
		t.Fatalf("Could not create new Collector: %v", err)
	}
	c.lastResult = nil
	c.nextSpeedtest = time.Now().Add(time.Hour)

	result := c.getSpeedtestResult()

	assert := assert.New(t)

	assert.NotEmpty(result)
	assert.Equal(mockSpeedtestResult, result)
	assert.Equal(result, c.lastResult)
	if !speedtestRan {
		t.Error("Speedtest was not called")
	}
}

func TestRunSpeedtestWhenCacheExpired(t *testing.T) {
	speedtestRan := false
	s := NewMockSpeedtest()
	s.Callback = func() {
		speedtestRan = true
	}

	c, err := NewCollector(defaultCacheTime, s)
	if err != nil {
		t.Fatalf("Could not create new Collector: %v", err)
	}
	c.lastResult = speedtest.NewFailedSpeedtestResult()
	c.nextSpeedtest = time.Now().Add(time.Hour * -1)

	result := c.getSpeedtestResult()

	assert := assert.New(t)

	assert.NotEmpty(result)
	assert.NotEqual(speedtest.NewFailedSpeedtestResult(), result)
	assert.Equal(mockSpeedtestResult, result)
	assert.Equal(result, c.lastResult)
	if !speedtestRan {
		t.Error("Speedtest was not called")
	}
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

	c, err := NewCollector(defaultCacheTime, s)
	if err != nil {
		t.Fatalf("Could not create new Collector: %v", err)
	}

	assert := assert.New(t)

	var result1, result2 *speedtest.SpeedtestResult
	go func() {
		result1 = c.getSpeedtestResult()
	}()
	<-sleeping
	assert.Nil(c.lastResult)
	result2 = c.getSpeedtestResult()

	assert.NotNil(result1)
	assert.Equal(result1, result2)
	assert.Equal(result1, c.lastResult)
	assert.Equal(1, i)
}

func TestCollect(t *testing.T) {
	s := NewMockSpeedtest()
	c, err := NewCollector(0, s)
	if err != nil {
		t.Fatalf("Could not create new Collector: %v", err)
	}

	t.Run("Success", func(t *testing.T) {
		ch := make(chan prometheus.Metric, 1)
		go c.Collect(ch)

		actualLabelValues := []string{mockSpeedtestResult.ClientIP(), mockSpeedtestResult.ClientISP()}

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
