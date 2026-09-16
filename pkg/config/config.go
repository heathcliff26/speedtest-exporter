package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/heathcliff26/promremote/v2/promremote"
	"go.yaml.in/yaml/v3"
)

const (
	DEFAULT_CONFIG_PATH           = "/etc/speedtest-exporter/config.yaml"
	DEFAULT_CONFIG_PATH_CONTAINER = "/config/config.yaml"

	DEFAULT_LOG_LEVEL       = "info"
	DEFAULT_PORT            = 8080
	DEFAULT_CACHE           = 5 * time.Minute
	DEFAULT_PERSIST_CACHE   = true
	DEFAULT_REMOTE_JOB_NAME = "speedtest-exporter"
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
	Instance     string        `yaml:"instance,omitempty"`
	Cache        time.Duration `yaml:"cache,omitempty"`
	PersistCache bool          `yaml:"persistCache,omitempty"`
	SpeedtestCLI string        `yaml:"speedtestCLI,omitempty"`
	ServerID     int           `yaml:"serverID,omitempty"`
	Remote       RemoteConfig  `yaml:"remote,omitempty"`
}

type RemoteConfig struct {
	Enable   bool   `yaml:"enable"`
	URL      string `yaml:"url"`
	Instance string `yaml:"instance,omitempty"`
	JobName  string `yaml:"jobName,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// Returns a Config with default values set
func DefaultConfig() Config {
	hostname, err := os.Hostname()
	if err != nil {
		slog.Error("Failed to retrieve hostname, using localhost instead", "err", err)
		hostname = "localhost"
	}
	return Config{
		LogLevel:     DEFAULT_LOG_LEVEL,
		Port:         DEFAULT_PORT,
		Instance:     hostname,
		Cache:        DEFAULT_CACHE,
		PersistCache: DEFAULT_PERSIST_CACHE,
		Remote: RemoteConfig{
			JobName: DEFAULT_REMOTE_JOB_NAME,
		},
	}
}

// Return the path to the config file
func getPath(path string) string {
	if path != "" {
		return path
	}
	if _, ok := os.LookupEnv("container"); ok {
		return DEFAULT_CONFIG_PATH_CONTAINER
	} else {
		return DEFAULT_CONFIG_PATH
	}
}

// Loads config from file, returns error if config is invalid
// Arguments:
//
//	path: Path to config file
//	mode: Mode used, determines how the config will be validated and which values will be processed
//	env: Determines if enviroment variables in the file will be expanded before decoding
func LoadConfig(path string, env bool) (Config, error) {
	c, err := loadConfigFile(path, env)
	if err != nil {
		return Config{}, err
	}

	err = setLogLevel(c.LogLevel)
	if err != nil {
		return Config{}, err
	}

	if c.Remote.Instance == "" {
		c.Remote.Instance = c.Instance
	}

	if c.Remote.Enable {
		if c.Remote.URL == "" {
			return Config{}, promremote.ErrMissingEndpoint{}
		}
		if c.Remote.Username != c.Remote.Password && (c.Remote.Username == "" || c.Remote.Password == "") {
			return Config{}, promremote.ErrMissingAuthCredentials{}
		}
	}

	return c, nil
}

func loadConfigFile(path string, env bool) (Config, error) {
	c := DefaultConfig()

	p := getPath(path)

	// #nosec G304 -- Local users can decide on their file path themselves.
	f, err := os.ReadFile(p)
	if path == "" && os.IsNotExist(err) {
		slog.Info("No config file specified and default file does not exist, falling back to default values.", slog.String("default-path", p))
		return c, nil
	} else if err != nil {
		return Config{}, err
	}

	if env {
		f = []byte(os.ExpandEnv(string(f)))
	}

	err = yaml.Unmarshal(f, &c)
	if err != nil {
		return Config{}, err
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
