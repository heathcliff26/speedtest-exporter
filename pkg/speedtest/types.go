package speedtest

type Speedtest interface {
	Speedtest() *SpeedtestResult
}

type SpeedtestResult struct {
	jitterLatency float64 // ms
	ping          float64 // ms
	downloadSpeed float64 // Mbit/s
	uploadSpeed   float64 // Mbit/s
	dataUsed      float64 // MB
	clientIsp     string
	clientIp      string
	success       bool
}

// Create a new SpeedtestResult for a failed speedtest.
func NewFailedSpeedtestResult() *SpeedtestResult {
	return &SpeedtestResult{
		success: false,
	}
}

// Create a new SpeedtestResult from a successfull speedrun
func NewSpeedtestResult(jitterLatency float64, ping float64, downloadSpeed float64, uploadSpeed float64, dataUsed float64, clientIsp string, clientIp string) *SpeedtestResult {
	return &SpeedtestResult{
		jitterLatency: jitterLatency,
		ping:          ping,
		downloadSpeed: downloadSpeed,
		uploadSpeed:   uploadSpeed,
		dataUsed:      dataUsed,
		clientIsp:     clientIsp,
		clientIp:      clientIp,
		success:       true,
	}
}

// Jitter latency of ping in ms
func (r *SpeedtestResult) JitterLatency() float64 {
	return r.jitterLatency
}

// Ping in ms
func (r *SpeedtestResult) Ping() float64 {
	return r.ping
}

// Download speed im Mbps
func (r *SpeedtestResult) DownloadSpeed() float64 {
	return r.downloadSpeed
}

// Upload speed in Mbps
func (r *SpeedtestResult) UploadSpeed() float64 {
	return r.uploadSpeed
}

// Data usage of speedtest in MB
func (r *SpeedtestResult) DataUsed() float64 {
	return r.dataUsed
}

// ISP name of the client/connection
func (r *SpeedtestResult) ClientIsp() string {
	return r.clientIsp
}

// Public IP of the client/connection
func (r *SpeedtestResult) ClientIp() string {
	return r.clientIp
}

// Indicates if the test was successfull
func (r *SpeedtestResult) Success() bool {
	return r.success
}
