package cfgsrv

type Message struct {
	OP      string      `json:"op"`
	Servers []string    `json:"servers,omitempty"`
	Config  interface{} `json:"config,omitempty"`
	Timeout string      `json:"timeout,omitempty"`
	Addr    string      `json:"addr,omitempty"`
}
