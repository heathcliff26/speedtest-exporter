package speedtest

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSpeedtestCLI(t *testing.T) {
	assert := assert.New(t)

	t.Run("Fail", func(t *testing.T) {
		_, err := NewSpeedtestCLI("/path/to/nothing")
		assert.Error(err)
		assert.ErrorIs(err, os.ErrNotExist)
	})

	t.Run("Success", func(t *testing.T) {
		s, err := NewSpeedtestCLI("testdata/speedtest-cli.sh")
		if err != nil {
			t.Fatalf("Failed to create speedtest-cli: %v", err)
		}
		assert.Contains(s.Path(), "testdata/speedtest-cli.sh")
	})
}

func TestRunSpeedtestForCLI(t *testing.T) {
	s, err := NewSpeedtestCLI("testdata/speedtest-cli.sh")
	if err != nil {
		t.Fatalf("Failed to create speedtest-cli: %v", err)
	}
	makeCmd = func(path string) *exec.Cmd {
		return exec.Command("bash", "-c", path)
	}

	expectedResult := NewSpeedtestResult(0.629, 17.148, 931.564032, 49.4518, 1141.3079899999998, "Some ISP", "100.107.156.96")

	result := s.Speedtest()

	assert.Equal(t, result, expectedResult)
}
