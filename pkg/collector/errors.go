package collector

type ErrNoSpeedtest struct{}

func (e ErrNoSpeedtest) Error() string {
	return "No valid speedtest provided"
}
