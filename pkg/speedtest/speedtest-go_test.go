package speedtest

import (
	"strconv"
	"testing"

	"github.com/showwin/speedtest-go/speedtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSpeedtestForGo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	s := NewSpeedtest(0)
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

func TestPickServer(t *testing.T) {
	client := speedtest.New()
	serverList, err := client.FetchServers()
	require.NoError(t, err, "Should fetch server list")
	require.NotEmpty(t, serverList, "Server list should not be empty")

	t.Run("Automatic", func(t *testing.T) {
		server := pickServer(serverList, 0)
		require.NotNil(t, server, "Should select a server automatically")
	})

	t.Run("Pinned", func(t *testing.T) {
		id, err := strconv.Atoi(serverList[0].ID)
		require.NoError(t, err, "Server ID should be numeric")

		pinned := pickServer(serverList, id)
		require.NotNil(t, pinned, "Should select the pinned server")
		assert.Equal(t, serverList[0].ID, pinned.ID, "Should use the pinned server")
	})

	t.Run("UnknownServer", func(t *testing.T) {
		server := pickServer(serverList, -1)
		require.NotNil(t, server, "Should fall back to automatic selection")
	})
}
