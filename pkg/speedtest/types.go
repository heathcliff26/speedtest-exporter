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
	serverID      string
	serverHost    string
	clientISP     string
	clientIP      string
	success       bool
}

// Create a new SpeedtestResult for a failed speedtest.
func NewFailedSpeedtestResult() *SpeedtestResult {
	return &SpeedtestResult{
		success: false,
	}
}

// Create a new SpeedtestResult from a successfull speedrun
func NewSpeedtestResult(jitterLatency, ping, downloadSpeed, uploadSpeed, dataUsed float64, serverID, serverHost, clientISP, clientIP string) *SpeedtestResult {
	return &SpeedtestResult{
		jitterLatency: jitterLatency,
		ping:          ping,
		downloadSpeed: downloadSpeed,
		uploadSpeed:   uploadSpeed,
		dataUsed:      dataUsed,
		serverID:      serverID,
		serverHost:    serverHost,
		clientISP:     clientISP,
		clientIP:      clientIP,
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

// ID of the speedtest server used for the test
func (r *SpeedtestResult) ServerID() string {
	return r.serverID
}

// Name of the speedtest server used for the test
func (r *SpeedtestResult) ServerHost() string {
	return r.serverHost
}

// ISP name of the client/connection
func (r *SpeedtestResult) ClientISP() string {
	return r.clientISP
}

// Public IP of the client/connection
func (r *SpeedtestResult) ClientIP() string {
	return r.clientIP
}

// Indicates if the test was successfull
func (r *SpeedtestResult) Success() bool {
	return r.success
}
