package cfgsrv

type Message struct {
	OP      string      `json:"op"`
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Servers []string    `json:"servers,omitempty"`
	Config  interface{} `json:"config,omitempty"`
	Timeout string      `json:"timeout,omitempty"`
	Addr    string      `json:"addr,omitempty"`
}

const (
	OPGet            = "get"
	OPConnect        = "connect"
	OPPing           = "ping"
	OPPong           = "pong"
	OPServersChanged = "servers_changed"
	OPConfigChanged  = "config_changed"

	TypeRequest  = "request"
	TypeResponse = "response"
	TypePush     = "push"
)
