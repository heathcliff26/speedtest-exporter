package version

import (
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	result := Version()

	lines := strings.Split(result, "\n")

	assert := assert.New(t)

	buildinfo, _ := debug.ReadBuildInfo()

	require.Equal(t, 5, len(lines), "Should have enough lines")
	assert.Contains(lines[0], Name)
	assert.Contains(lines[1], buildinfo.Main.Version)

	commit := strings.Split(lines[2], ":")
	assert.NotEmpty(strings.TrimSpace(commit[1]))

	assert.Contains(lines[3], runtime.Version())

	assert.Equal("", lines[4], "Should have trailing newline")
}
