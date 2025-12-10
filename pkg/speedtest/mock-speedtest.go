package speedtest

import "time"

type MockSpeedtest struct {
	Callback func()
	Fail     bool
	Result   *SpeedtestResult
}

func (s *MockSpeedtest) Speedtest() *SpeedtestResult {
	if s.Callback != nil {
		s.Callback()
	}
	if s.Fail {
		return NewFailedSpeedtestResult()
	}
	return s.Result
}

func MockSpeedtestResult(timestamp int64) *SpeedtestResult {
	result := NewSpeedtestResult(0.5, 15, 876.53, 12.34, 950.3079, "1234", "example.org", "Foo Corp.", "127.0.0.1", 251234*time.Millisecond)
	result.timestamp = timestamp
	return result
}
