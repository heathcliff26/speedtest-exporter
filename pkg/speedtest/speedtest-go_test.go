package speedtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunSpeedtestForGo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	s := NewSpeedtest()
	result := s.Speedtest()
	if !result.Success() {
		t.Fatal("Speedtest returned with failure")
	}

	assert := assert.New(t)
	assert.NotEmpty(result)
	assert.NotEmpty(result.DownloadSpeed())
	assert.NotEmpty(result.UploadSpeed())
	assert.NotEmpty(result.DataUsed())
	assert.NotEmpty(result.ClientIsp())
	assert.NotEmpty(result.ClientIp())
}
