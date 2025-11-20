package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/heathcliff26/speedtest-exporter/pkg/collector"
	"github.com/heathcliff26/speedtest-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerRootHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err, "Failed to create request")
	rr := httptest.NewRecorder()

	ServerRootHandler(rr, req)

	assert := assert.New(t)

	assert.Equal(http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(body, "<html>")
	assert.Contains(body, "</html>")
	assert.Contains(body, "<a href='/metrics'>")
}

func TestCreateSpeedtest(t *testing.T) {
	t.Run("SpeedtestCLI", func(t *testing.T) {
		s, err := createSpeedtest("../pkg/speedtest/testdata/speedtest-cli.sh")
		require.NoError(t, err, "Should create speedtest-cli")
		assert.Equal(t, "*speedtest.SpeedtestCLI", reflect.TypeOf(s).String())
	})
	t.Run("Speedtest", func(t *testing.T) {
		s, err := createSpeedtest("")
		require.NoError(t, err, "Should create speedtest-cli")
		assert.Equal(t, "*speedtest.SpeedtestGo", reflect.TypeOf(s).String())
	})
}

func TestServerWriteTimeout(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s, err := createSpeedtest("")
	require.NoError(err, "Should create speedtest")
	c, err := collector.NewCollector(nil, s, "testinstance") // Ensure we do not use a cache
	require.NoError(err, "Should create collector")

	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	server := createServer(config.DEFAULT_PORT, reg) // Use port 0 to let the OS assign a free port
	require.NotNil(server, "Server should not be nil")

	serverError := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
			err = nil
		}
		serverError <- err
	}()

	addr := fmt.Sprintf("http://localhost:%d", config.DEFAULT_PORT)

	require.Eventually(func() bool {
		res, err := http.Get(addr)
		if err == nil && res.StatusCode == http.StatusOK {
			return true
		}
		return false
	}, 1*time.Minute, 10*time.Second, "Server should start within 1 minute")

	res, err := http.Get(addr + "/metrics")
	require.NoError(err, "Should be able to get metrics")
	assert.Equal(http.StatusOK, res.StatusCode, "Should return status OK")

	assert.NoError(server.Shutdown(t.Context()), "Server should shut down without error")
	assert.NoError(<-serverError, "Server should not return an error on shutdown")
}
