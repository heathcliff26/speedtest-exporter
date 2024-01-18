package speedtest

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"os/exec"
)

type SpeedtestCLI struct {
	path string
}

// Create SpeedtestCLI, fails when it can't find the speedtest-cli binary
// Arguments:
//
//	executable: name or full path to speedtest-cli binary
func NewSpeedtestCLI(executable string) (*SpeedtestCLI, error) {
	path, err := exec.LookPath(executable)
	if errors.Is(err, exec.ErrDot) {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	return &SpeedtestCLI{
		path: path,
	}, nil
}

// Get path of the speedtest-cli binary
func (s *SpeedtestCLI) Path() string {
	return s.path
}

var makeCmd = func(path string) *exec.Cmd {
	return exec.Command(path, "--format=json-pretty", "--accept-license", "--accept-gdpr")
}

// Execute the speedtest-cli binary and parse the result
func (s *SpeedtestCLI) Speedtest() *SpeedtestResult {
	cmd := makeCmd(s.Path())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		slog.Error("Could not execute speedtest", "error", err, slog.String("stdout", stdout.String()), slog.String("stderr", stderr.String()))
		return NewFailedSpeedtestResult()
	}

	var out resultJSON
	err = json.Unmarshal(stdout.Bytes(), &out)
	if err != nil {
		slog.Error("Parsing JSON output from speedtest failed", "error", err, slog.String("output", stdout.String()))
		return NewFailedSpeedtestResult()
	}

	downloadMbps := convertBytesToMbits(out.Download.Bandwidth)
	uploadMbps := convertBytesToMbits(out.Upload.Bandwidth)
	dataUsed := convertBytesToMB(out.Download.Bytes) + convertBytesToMB(out.Upload.Bytes)

	slog.Info("Successfully ran speedtest", slog.Group("result"),
		slog.Float64("jitterLatency", out.Ping.Jitter),
		slog.Float64("ping", out.Ping.Latency),
		slog.Float64("downloadSpeed", downloadMbps),
		slog.Float64("uploadSpeed", uploadMbps),
		slog.Float64("dataUsed", dataUsed),
		slog.Int("serverId", out.Server.Id),
		slog.String("serverHost", out.Server.Host),
		slog.String("isp", out.ISP),
		slog.String("IP", out.Interface.ExternalIP),
	)

	return NewSpeedtestResult(out.Ping.Jitter, out.Ping.Latency, downloadMbps, uploadMbps, dataUsed, out.ISP, out.Interface.ExternalIP)
}
