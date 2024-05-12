package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/heathcliff26/promremote/promremote"
	"github.com/heathcliff26/speedtest-exporter/pkg/collector"
	"github.com/heathcliff26/speedtest-exporter/pkg/config"
	"github.com/heathcliff26/speedtest-exporter/pkg/speedtest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	configPath string
	env        bool
)

// Initialize the logger
func init() {
	flag.StringVar(&configPath, "config", "", "Optional: Path to config file")
	flag.BoolVar(&env, "env", false, "Used together with -config, when set will expand enviroment variables in config")
}

// Handle requests to the webroot.
// Serves static, human-readable HTML that provides a link to /metrics
func ServerRootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<html><body><h1>Welcome to speedtest-exporter</h1>Click <a href='/metrics'>here</a> to see metrics.</body></html>")
}

func createSpeedtest(path string) (speedtest.Speedtest, error) {
	if path == "" {
		slog.Debug("Using go-native speedtest implementation")
		return speedtest.NewSpeedtest(), nil
	} else {
		slog.Debug("Using external speedtest-cli binary", "path", path)
		return speedtest.NewSpeedtestCLI(path)
	}
}

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(configPath, env)
	if err != nil {
		slog.Error("Could not load configuration", slog.String("path", configPath), slog.String("err", err.Error()))
		os.Exit(1)
	}

	s, err := createSpeedtest(cfg.SpeedtestCLI)
	if err != nil {
		slog.Error("Failed initialize speedtest", "err", err)
		os.Exit(1)
	}

	collector, err := collector.NewCollector(cfg.Cache, s)
	if err != nil {
		slog.Error("Failed to create collector", "err", err)
		os.Exit(1)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(collector)

	if cfg.Remote.Enable {
		rwClient, err := promremote.NewWriteClient(cfg.Remote.URL, cfg.Remote.Instance, "integrations/speedtest", reg)
		if err != nil {
			slog.Error("Failed to create remote write client", "err", err)
			os.Exit(1)
		}
		if cfg.Remote.Username != "" {
			err := rwClient.SetBasicAuth(cfg.Remote.Username, cfg.Remote.Password)
			if err != nil {
				slog.Error("Failed to create remote_write client", "err", err)
				os.Exit(1)
			}
		}

		slog.Info("Starting remote_write client", slog.String("interval", cfg.Cache.String()))
		rwQuit := make(chan bool)
		rwClient.Run(cfg.Cache, rwQuit)
		defer func() {
			rwQuit <- true
			close(rwQuit)
		}()
	}

	http.HandleFunc("/", ServerRootHandler)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	addr := ":" + strconv.Itoa(cfg.Port)
	slog.Info("Starting http server", slog.String("addr", addr))
	err = http.ListenAndServe(addr, nil)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Failed to start http server", "err", err)
		os.Exit(1)
	}
}
