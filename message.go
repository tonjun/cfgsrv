package cfgsrv

type Message struct {
	OP      string      `json:"op"`
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Peers   []string    `json:"peers,omitempty"`
	Config  interface{} `json:"config,omitempty"`
	Timeout string      `json:"timeout,omitempty"`
	Addr    string      `json:"addr,omitempty"`
}

const (
	OPGet           = "get"
	OPConnect       = "connect"
	OPPing          = "ping"
	OPPong          = "pong"
	OPPeersChanged  = "peers_changed"
	OPConfigChanged = "config_changed"

	TypeRequest  = "request"
	TypeResponse = "response"
	TypePush     = "push"
)
