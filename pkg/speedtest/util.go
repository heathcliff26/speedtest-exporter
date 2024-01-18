package speedtest

// Convert unit bytes to megabits
func convertBytesToMbits(bytes int64) float64 {
	return convertBytesToMB(bytes) * 8
}

// Convert unit bytes to megabytes
func convertBytesToMB(bytes int64) float64 {
	return float64(bytes) / 1000000
}
