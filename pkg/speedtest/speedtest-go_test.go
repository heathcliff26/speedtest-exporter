package speedtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSpeedtestForGo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	s := NewSpeedtest()
	result := s.Speedtest()
	require.True(t, result.Success(), "Speedtest should succeed")

	assert := assert.New(t)
	assert.NotEmpty(result)
	assert.NotEmpty(result.DownloadSpeed())
	assert.NotEmpty(result.UploadSpeed())
	assert.NotEmpty(result.DataUsed())
	assert.NotEmpty(result.ClientISP())
	assert.NotEmpty(result.ClientIP())
}
