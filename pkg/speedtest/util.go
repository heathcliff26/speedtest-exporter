package speedtest

import (
	"log/slog"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | speedtest.ByteRate
}

// Convert unit bytes to megabits
func convertBytesToMbits[T Number](bytes T) float64 {
	return convertBytesToMB(bytes) * 8
}

// Convert unit bytes to megabytes
func convertBytesToMB[T Number](bytes T) float64 {
	return float64(bytes) / speedtest.MB
}

// Print the log message for a successful speedtest
func printSuccessMessage(res *SpeedtestResult) {
	slog.Info("Successfully ran speedtest",
		slog.Float64("jitterLatency", res.JitterLatency()),
		slog.Float64("ping", res.Ping()),
		slog.Float64("downloadSpeed", res.DownloadSpeed()),
		slog.Float64("uploadSpeed", res.UploadSpeed()),
		slog.Float64("dataUsed", res.DataUsed()),
		slog.String("serverID", res.ServerID()),
		slog.String("serverHost", res.ServerHost()),
		slog.String("isp", res.ClientISP()),
		slog.String("ip", res.ClientIP()),
		slog.Duration("duration", time.Duration(res.Duration())*time.Millisecond),
	)
}
