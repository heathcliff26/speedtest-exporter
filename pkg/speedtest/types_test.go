package speedtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFailedSpeedtestResult(t *testing.T) {
	var expectedResult = &SpeedtestResult{
		success: false,
	}
	assert.Equal(t, expectedResult, NewFailedSpeedtestResult())
}

func TestNewSpeedtestResult(t *testing.T) {
	var expectedResult = &SpeedtestResult{
		jitterLatency: 0.5,
		ping:          15,
		downloadSpeed: 876.53,
		uploadSpeed:   12.34,
		dataUsed:      950.3079,
		clientIsp:     "Foo Corp.",
		clientIp:      "127.0.0.1",
		success:       true,
	}
	actualResult := NewSpeedtestResult(0.5, 15, 876.53, 12.34, 950.3079, "Foo Corp.", "127.0.0.1")
	assert.Equal(t, expectedResult, actualResult)
}
