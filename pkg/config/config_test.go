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
	c1 := Config{
		LogLevel:     "warn",
		Port:         80,
		Cache:        Duration(time.Minute),
		SpeedtestCLI: "/path/to/speedtest",
	}
	c2 := Config{
		LogLevel: "debug",
		Port:     2080,
		Cache:    Duration(30 * time.Minute),
		Remote: RemoteConfig{
			Enable:   true,
			URL:      "https://example.org/",
			Instance: "test",
			Username: "somebody",
			Password: "somebody's password",
		},
	}
	c3 := Config{
		LogLevel: "error",
		Port:     DEFAULT_PORT,
		Cache:    DEFAULT_CACHE,
		Remote: RemoteConfig{
			Enable:   true,
			URL:      "https://example.org/",
			Instance: "test",
		},
	}
	tMatrix := []struct {
		Name, Path string
		Result     Config
	}{
		{
			Name:   "EmptyConfig",
			Path:   "",
			Result: DefaultConfig(),
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
			c, err := LoadConfig(tCase.Path, false)

			require.NoError(t, err, "Should load config")
			assert.Equal(t, tCase.Result, c)
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
			Error: "*fmt.wrapError",
		},
		{
			Name:  "InvalidCache",
			Path:  "testdata/invalid-config-1.yaml",
			Error: "*fmt.wrapError",
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
	c := Config{
		LogLevel: "debug",
		Port:     2080,
		Cache:    Duration(time.Minute),
	}
	t.Setenv("SPEEDTEST_TEST_LOG_LEVEL", c.LogLevel)
	t.Setenv("SPEEDTEST_TEST_PORT", strconv.Itoa(c.Port))
	t.Setenv("SPEEDTEST_TEST_CACHE", c.Cache.String())

	res, err := LoadConfig("testdata/env-config.yaml", true)

	assert := assert.New(t)

	assert.NoError(err)
	assert.Equal(c, res)
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
			t.Fatalf("Failed to cleanup after test: %v", err)
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
