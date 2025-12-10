package speedtest

import (
	"encoding/json"
	"time"
)

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
	timestamp     int64 // milliseconds since Unix epoch
	duration      int64 // milliseconds
}

// Create a new SpeedtestResult for a failed speedtest.
func NewFailedSpeedtestResult() *SpeedtestResult {
	return &SpeedtestResult{
		success:   false,
		timestamp: time.Now().UnixMilli(),
	}
}

// Create a new SpeedtestResult from a successful speedrun
func NewSpeedtestResult(jitterLatency, ping, downloadSpeed, uploadSpeed, dataUsed float64, serverID, serverHost, clientISP, clientIP string, duration time.Duration) *SpeedtestResult {
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
		timestamp:     time.Now().UnixMilli(),
		duration:      duration.Milliseconds(),
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

// Download speed in Mbps
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

// Indicates if the test was successful
func (r *SpeedtestResult) Success() bool {
	return r.success
}

// Returns the timestamp of when the speedtest was run.
// The timestamp is represented as milliseconds since the Unix epoch.
func (r *SpeedtestResult) Timestamp() int64 {
	return r.timestamp
}

// Returns the timestamp of when the speedtest was run as time.Time
func (r *SpeedtestResult) TimestampAsTime() time.Time {
	return time.UnixMilli(r.timestamp)
}

// Duration of the speedtest in milliseconds
func (r *SpeedtestResult) Duration() int64 {
	return r.duration
}

type speedtestResultJSONAlias struct {
	JitterLatency float64 `json:"jitter_latency_ms"`
	Ping          float64 `json:"ping_ms"`
	DownloadSpeed float64 `json:"download_mbps"`
	UploadSpeed   float64 `json:"upload_mbps"`
	DataUsed      float64 `json:"data_used_mb"`
	ServerID      string  `json:"server_id"`
	ServerHost    string  `json:"server_host"`
	ClientISP     string  `json:"client_isp"`
	ClientIP      string  `json:"client_ip"`
	Success       bool    `json:"success"`
	Timestamp     int64   `json:"timestamp"`
	Duration      int64   `json:"duration_ms"`
}

// MarshalJSON implements json.Marshaler so the (unexported) fields of
// SpeedtestResult can be serialized with meaningful JSON keys.
func (r *SpeedtestResult) MarshalJSON() ([]byte, error) {
	a := speedtestResultJSONAlias{
		JitterLatency: r.jitterLatency,
		Ping:          r.ping,
		DownloadSpeed: r.downloadSpeed,
		UploadSpeed:   r.uploadSpeed,
		DataUsed:      r.dataUsed,
		ServerID:      r.serverID,
		ServerHost:    r.serverHost,
		ClientISP:     r.clientISP,
		ClientIP:      r.clientIP,
		Success:       r.success,
		Timestamp:     r.timestamp,
		Duration:      r.duration,
	}

	return json.MarshalIndent(a, "", "  ")
}

// UnmarshalJSON implements json.Unmarshaler so JSON with the same keys used
// in MarshalJSON can be decoded back into a SpeedtestResult with unexported fields.
func (r *SpeedtestResult) UnmarshalJSON(data []byte) error {
	var a speedtestResultJSONAlias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	r.jitterLatency = a.JitterLatency
	r.ping = a.Ping
	r.downloadSpeed = a.DownloadSpeed
	r.uploadSpeed = a.UploadSpeed
	r.dataUsed = a.DataUsed
	r.serverID = a.ServerID
	r.serverHost = a.ServerHost
	r.clientISP = a.ClientISP
	r.clientIP = a.ClientIP
	r.success = a.Success
	r.timestamp = a.Timestamp
	r.duration = a.Duration

	return nil
}
