package config

import (
	"log/slog"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidConfigs(t *testing.T) {
	defaultConfig := DefaultConfig()
	defaultConfig.Remote.Instance = defaultConfig.Instance
	c1 := Config{
		LogLevel:     "warn",
		Port:         80,
		Instance:     "test",
		Cache:        time.Minute,
		PersistCache: false,
		SpeedtestCLI: "/path/to/speedtest",
		ServerID:     12345,
		Remote: RemoteConfig{
			JobName:  DEFAULT_REMOTE_JOB_NAME,
			Instance: "test",
		},
	}
	c2 := Config{
		LogLevel:     "debug",
		Port:         2080,
		Instance:     "test",
		Cache:        30 * time.Minute,
		PersistCache: true,
		Remote: RemoteConfig{
			Enable:   true,
			URL:      "https://example.org/",
			Instance: "test",
			JobName:  "testjob",
			Username: "somebody",
			Password: "somebody's password",
		},
	}
	c3 := Config{
		LogLevel:     "error",
		Port:         DEFAULT_PORT,
		Instance:     "another-instance",
		Cache:        DEFAULT_CACHE,
		PersistCache: DEFAULT_PERSIST_CACHE,
		Remote: RemoteConfig{
			Enable:   true,
			URL:      "https://example.org/",
			Instance: "test",
			JobName:  DEFAULT_REMOTE_JOB_NAME,
		},
	}
	tMatrix := []struct {
		Name, Path string
		Result     Config
	}{
		{
			Name:   "EmptyConfig",
			Path:   "",
			Result: defaultConfig,
		},
		{
			Name:   "Config1",
			Path:   "testdata/valid-config-1.yaml",
			Result: c1,
		},
		{
			Name:   "Config2",
			Path:   "testdata/valid-config-2.yaml",
			Result: c2,
		},
		{
			Name:   "Config3",
			Path:   "testdata/valid-config-3.yaml",
			Result: c3,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			assert := assert.New(t)
			c, err := LoadConfig(tCase.Path, false)

			require.NoError(t, err, "Should load config")
			assert.Equal(tCase.Result, c)
		})
	}
}

func TestInvalidConfig(t *testing.T) {
	tMatrix := []struct {
		Name, Path, Mode, Error string
	}{
		{
			Name:  "InvalidPath",
			Path:  "file-does-not-exist.yaml",
			Error: "*fs.PathError",
		},
		{
			Name:  "NotYaml",
			Path:  "testdata/not-a-config.txt",
			Error: "*yaml.TypeError",
		},
		{
			Name:  "InvalidCache",
			Path:  "testdata/invalid-config-1.yaml",
			Error: "*yaml.TypeError",
		},
		{
			Name:  "MissingRemoteEndpoint",
			Path:  "testdata/invalid-config-2.yaml",
			Error: "promremote.ErrMissingEndpoint",
		},
		{
			Name:  "IncompleteRemoteCredentials",
			Path:  "testdata/invalid-config-3.yaml",
			Error: "promremote.ErrMissingAuthCredentials",
		},
		{
			Name:  "InvalidLogLevel",
			Path:  "testdata/invalid-loglevel.yaml",
			Error: "*config.ErrUnknownLogLevel",
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			require := require.New(t)

			_, err := LoadConfig(tCase.Path, false)

			require.Error(err, "Should return an error")
			require.Equal(tCase.Error, reflect.TypeOf(err).String(), "Should receive the expected error")
		})
	}
}

func TestEnvSubstitution(t *testing.T) {
	c := DefaultConfig()
	c.LogLevel = "debug"
	c.Port = 2080
	c.Cache = time.Minute
	c.PersistCache = DEFAULT_PERSIST_CACHE
	c.Remote.Instance = c.Instance

	t.Setenv("SPEEDTEST_TEST_LOG_LEVEL", c.LogLevel)
	t.Setenv("SPEEDTEST_TEST_PORT", strconv.Itoa(c.Port))
	t.Setenv("SPEEDTEST_TEST_CACHE", c.Cache.String())

	res, err := LoadConfig("testdata/env-config.yaml", true)

	assert := assert.New(t)

	assert.NotEmpty(res.Instance, "Should initialize instance from hostname")
	assert.Equal(res.Instance, res.Remote.Instance, "Should initialize remote instance from instance")

	assert.NoError(err)
	assert.Equal(c, res)
}

func TestGetPath(t *testing.T) {
	tMatrix := []struct {
		Name, Path string
		Container  bool
		Result     string
	}{
		{
			Name:   "GivenPath",
			Path:   "testpath",
			Result: "testpath",
		},
		{
			Name:   "Default",
			Result: DEFAULT_CONFIG_PATH,
		},
		{
			Name:      "Container",
			Container: true,
			Result:    DEFAULT_CONFIG_PATH_CONTAINER,
		},
	}

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			if tCase.Container {
				t.Setenv("container", "podman")
			}
			assert.Equal(t, tCase.Result, getPath(tCase.Path), "Should return the correct path")
		})
	}
}

func TestSetLogLevel(t *testing.T) {
	tMatrix := []struct {
		Name  string
		Level slog.Level
		Error error
	}{
		{"debug", slog.LevelDebug, nil},
		{"info", slog.LevelInfo, nil},
		{"warn", slog.LevelWarn, nil},
		{"error", slog.LevelError, nil},
		{"DEBUG", slog.LevelDebug, nil},
		{"INFO", slog.LevelInfo, nil},
		{"WARN", slog.LevelWarn, nil},
		{"ERROR", slog.LevelError, nil},
		{"Unknown", 0, &ErrUnknownLogLevel{"Unknown"}},
	}
	t.Cleanup(func() {
		err := setLogLevel(DEFAULT_LOG_LEVEL)
		if err != nil {
			t.Logf("Failed to cleanup after test: %v", err)
		}
	})

	for _, tCase := range tMatrix {
		t.Run(tCase.Name, func(t *testing.T) {
			err := setLogLevel(tCase.Name)

			require.Equal(t, tCase.Error, err, "Should return the expected error")
			if err == nil {
				assert.Equal(t, tCase.Level, logLevel.Level())
			}
		})
	}
}
