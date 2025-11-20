package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/heathcliff26/promremote/promremote"
	"github.com/heathcliff26/speedtest-exporter/pkg/cache"
	"github.com/heathcliff26/speedtest-exporter/pkg/collector"
	"github.com/heathcliff26/speedtest-exporter/pkg/config"
	"github.com/heathcliff26/speedtest-exporter/pkg/speedtest"
	"github.com/heathcliff26/speedtest-exporter/pkg/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	configPath  string
	env         bool
	showVersion bool
)

// Initialize the logger
func init() {
	flag.StringVar(&configPath, "config", "", "Optional: Path to config file")
	flag.BoolVar(&env, "env", false, "Used together with -config, when set will expand enviroment variables in config")
	flag.BoolVar(&showVersion, "version", false, "Show the version information and exit")
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

func createServer(port int, reg *prometheus.Registry) *http.Server {
	router := http.NewServeMux()
	router.HandleFunc("/", ServerRootHandler)
	router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	return &http.Server{
		Addr:        ":" + strconv.Itoa(port),
		Handler:     router,
		ReadTimeout: 10 * time.Second,
		// The speedtest takes roughly 22-24 seconds on average.
		// Ensure timeout has some buffer for a worst case.
		WriteTimeout: 60 * time.Second,
	}
}

func main() {
	flag.Parse()

	if showVersion {
		fmt.Print(version.Version())
		os.Exit(0)
	}

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

	resultCache := cache.NewCache(cfg.PersistCache, "/cache/speedtest-result.json", time.Duration(cfg.Cache))

	collector, err := collector.NewCollector(resultCache, s)
	if err != nil {
		slog.Error("Failed to create collector", "err", err)
		os.Exit(1)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(collector)

	if cfg.Remote.Enable {
		rwClient, err := promremote.NewWriteClient(cfg.Remote.URL, cfg.Remote.Instance, cfg.Remote.JobName, reg)
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
		rwClient.Run(time.Duration(cfg.Cache), rwQuit)
		defer func() {
			rwQuit <- true
			close(rwQuit)
		}()
	}

	server := createServer(cfg.Port, reg)

	slog.Info("Starting http server", slog.String("addr", server.Addr))
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("Failed to start http server", "err", err)
		os.Exit(1)
	}
}
