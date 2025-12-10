package speedtest

import (
	"log/slog"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

type SpeedtestGo struct {
}

// Create instance of Speedtest
func NewSpeedtest() *SpeedtestGo {
	return &SpeedtestGo{}
}

// Use the speedtest-go api to run a speedtest and parse the result
func (s *SpeedtestGo) Speedtest() *SpeedtestResult {
	start := time.Now()

	client := speedtest.New()

	serverList, err := client.FetchServers()
	if err != nil {
		slog.Error("Could not fetch server list", "error", err)
		return NewFailedSpeedtestResult()
	}
	targets, err := serverList.FindServer([]int{})
	if err != nil {
		slog.Error("Failed to find closest server", "error", err)
		return NewFailedSpeedtestResult()
	}
	if len(targets) != 1 {
		slog.Error("FindServer returned more than one server")
		return NewFailedSpeedtestResult()
	}
	server := targets[0]

	err = server.TestAll()
	if err != nil {
		slog.Error("Failed to run speedtest", "error", err)
		return NewFailedSpeedtestResult()
	}
	user, err := client.FetchUserInfo()
	if err != nil {
		slog.Error("Failed to fetch client information", "error", err)
		return NewFailedSpeedtestResult()
	}

	downloadMbps := convertBytesToMbits(server.DLSpeed)
	uploadMbps := convertBytesToMbits(server.ULSpeed)
	dataUsed := convertBytesToMB(server.Context.GetTotalDownload()) + convertBytesToMB(server.Context.GetTotalUpload())

	res := NewSpeedtestResult(float64(server.Jitter.Milliseconds()), float64(server.Latency.Milliseconds()), downloadMbps, uploadMbps, dataUsed, server.ID, server.Host, user.Isp, user.IP, time.Since(start))

	printSuccessMessage(res)

	return res
}
