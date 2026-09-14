package speedtest

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

type SpeedtestGo struct {
	serverID int
}

// Create instance of Speedtest. When serverID is 0, the closest server is chosen automatically.
func NewSpeedtest(serverID int) *SpeedtestGo {
	return &SpeedtestGo{
		serverID: serverID,
	}
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

	server := pickServer(serverList, s.serverID)
	if server == nil {
		return NewFailedSpeedtestResult()
	}

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

// Fetch the list of available Ookla speedtest servers, sorted by proximity.
func ListServers() (speedtest.Servers, error) {
	client := speedtest.New()
	return client.FetchServers()
}

// Select a server from the list, optionally pinned to serverID. When serverID is 0, the closest server is
// chosen automatically. Returns nil if no single server could be selected, having already logged the error.
func pickServer(serverList speedtest.Servers, serverID int) *speedtest.Server {
	ids := []int{}
	if serverID != 0 {
		ids = append(ids, serverID)
	}
	targets, err := serverList.FindServer(ids)
	if err != nil {
		slog.Error("Failed to find closest server", "error", err)
		return nil
	}
	if len(targets) != 1 {
		slog.Error("FindServer returned more than one server")
		return nil
	}
	server := targets[0]

	if serverID != 0 && server.ID != strconv.Itoa(serverID) {
		slog.Warn("Pinned server not found in server list, falling back to automatic selection", "pinnedServerID", serverID, "selectedServerID", server.ID)
	}

	return server
}
