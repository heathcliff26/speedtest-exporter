package speedtest

import (
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
