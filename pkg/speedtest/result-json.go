package speedtest

// Data structure for the json output of speedtest-cli
type resultJSON struct {
	Type       string              `json:"type"`
	Timestamp  string              `json:"timestamp"`
	Ping       resultPingJSON      `json:"ping"`
	Download   resultBandwidthJSON `json:"download"`
	Upload     resultBandwidthJSON `json:"upload"`
	PacketLoss float64             `json:"packetLoss"`
	ISP        string              `json:"isp"`
	Interface  resultInterfaceJSON `json:"interface"`
	Server     resultServerJSON    `json:"server"`
	Result     interface{}         `json:"result"`
}

type resultPingJSON struct {
	Jitter  float64 `json:"jitter"`
	Latency float64 `json:"latency"`
	Low     float64 `json:"low"`
	High    float64 `json:"high"`
}

type resultBandwidthJSON struct {
	// Unit: Bytes
	Bandwidth int64       `json:"bandwidth"`
	Bytes     int64       `json:"bytes"`
	Elapsed   int64       `json:"elapsed"`
	Latency   interface{} `json:"latency"`
}

type resultInterfaceJSON struct {
	InternalIP string `json:"internalIp"`
	Name       string `json:"name"`
	MACAddr    string `json:"macAddr"`
	IsVPN      bool   `json:"isVPN"`
	ExternalIP string `json:"externalIp"`
}

type resultServerJSON struct {
	Id       int    `json:"id"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Country  string `json:"country"`
	IP       string `json:"ip"`
}
