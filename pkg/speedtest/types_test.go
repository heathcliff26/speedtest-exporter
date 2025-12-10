package speedtest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFailedSpeedtestResult(t *testing.T) {
	assert := assert.New(t)
	var expectedResult = &SpeedtestResult{
		success: false,
	}

	result := NewFailedSpeedtestResult()
	assert.NotZero(result.Timestamp(), "Failed result should have timestamp")
	expectedResult.timestamp = result.timestamp // align timestamps for comparison
	assert.Equal(expectedResult, result, "Should match expected failed SpeedtestResult")
}

func TestNewSpeedtestResult(t *testing.T) {
	assert := assert.New(t)

	var expectedResult = &SpeedtestResult{
		jitterLatency: 0.5,
		ping:          15,
		downloadSpeed: 876.53,
		uploadSpeed:   12.34,
		dataUsed:      950.3079,
		serverID:      "1234",
		serverHost:    "example.org",
		clientISP:     "Foo Corp.",
		clientIP:      "127.0.0.1",
		success:       true,
		duration:      231234,
	}

	actualResult := NewSpeedtestResult(0.5, 15, 876.53, 12.34, 950.3079, "1234", "example.org", "Foo Corp.", "127.0.0.1", 231234*time.Millisecond)

	assert.InDelta(time.Now().UnixMilli(), actualResult.Timestamp(), 10, "Timestamp should be close to current time")
	expectedResult.timestamp = actualResult.timestamp // align timestamps for comparison
	assert.Equal(expectedResult, actualResult, "NewSpeedtestResult should create the expected SpeedtestResult")
}

func TestSpeedtestResultJSON(t *testing.T) {
	assert := assert.New(t)

	result := MockSpeedtestResult(1234)

	jsonData, err := result.MarshalJSON()
	assert.NoError(err, "Should marshal SpeedtestResult to JSON without error")

	unmarshaledResult := &SpeedtestResult{}
	err = unmarshaledResult.UnmarshalJSON(jsonData)
	assert.NoError(err, "Should unmarshal JSON to SpeedtestResult without error")

	assert.Equal(result, unmarshaledResult, "Unmarshaled result should match the original")

	failedMarshal := &SpeedtestResult{}
	err = failedMarshal.UnmarshalJSON([]byte("not-valid-json"))
	assert.Error(err, "Should return error when unmarshaling invalid JSON")
	assert.Empty(failedMarshal, "Should not change anything in the target variable")
}

func TestSpeedtestResultTimestampAsTime(t *testing.T) {
	r := MockSpeedtestResult(123456789)

	assert.Equal(t, time.UnixMilli(r.timestamp), r.TimestampAsTime(), "TimestampAsTime should return correct time.Time representation")
}
