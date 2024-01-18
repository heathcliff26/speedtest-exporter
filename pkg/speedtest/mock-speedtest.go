package speedtest

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
