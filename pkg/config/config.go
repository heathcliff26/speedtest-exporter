package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/heathcliff26/promremote/promremote"
	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_LOG_LEVEL = "info"
	DEFAULT_PORT      = 8080
	DEFAULT_CACHE     = time.Duration(5 * time.Minute)
)

var logLevel *slog.LevelVar

// Initialize the logger
func init() {
	logLevel = &slog.LevelVar{}
	opts := slog.HandlerOptions{
		Level: logLevel,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
}

type Config struct {
	LogLevel     string        `yaml:"logLevel,omitempty"`
	Port         int           `yaml:"port,omitempty"`
	Cache        time.Duration `yaml:"cache,omitempty"`
	SpeedtestCLI string        `yaml:"speedtestCLI,omitempty"`
	Remote       RemoteConfig  `yaml:"remote,omitempty"`
}

type RemoteConfig struct {
	Enable   bool   `yaml:"enable"`
	URL      string `yaml:"url"`
	Instance string `yaml:"instance"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// Returns a Config with default values set
func DefaultConfig() Config {
	return Config{
		LogLevel: DEFAULT_LOG_LEVEL,
		Port:     DEFAULT_PORT,
		Cache:    DEFAULT_CACHE,
	}
}

// Loads config from file, returns error if config is invalid
// Arguments:
//
//	path: Path to config file
//	mode: Mode used, determines how the config will be validated and which values will be processed
//	env: Determines if enviroment variables in the file will be expanded before decoding
func LoadConfig(path string, env bool) (Config, error) {
	c := DefaultConfig()

	if path == "" {
		_ = setLogLevel(DEFAULT_LOG_LEVEL)
		return c, nil
	}

	f, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	if env {
		f = []byte(os.ExpandEnv(string(f)))
	}

	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return Config{}, err
	}

	err = setLogLevel(c.LogLevel)
	if err != nil {
		return Config{}, err
	}

	if c.Remote.Enable {
		if c.Remote.URL == "" {
			return Config{}, promremote.ErrMissingEndpoint{}
		}
		if c.Remote.Username != c.Remote.Password && (c.Remote.Username == "" || c.Remote.Password == "") {
			return Config{}, promremote.ErrMissingAuthCredentials{}
		}
		if c.Remote.Instance == "" {
			slog.Info("No instance name provided, defaulting to hostname")
			hostname, err := os.Hostname()
			if err != nil {
				slog.Error("Failed to retrieve hostname, using localhost instead", "err", err)
				hostname = "localhost"
			}
			c.Remote.Instance = hostname
		}
	}

	return c, nil
}

// Parse a given string and set the resulting log level
func setLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		return &ErrUnknownLogLevel{level}
	}
	return nil
}
