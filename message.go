package cfgsrv

import (
	"encoding/json"
)

// Message is the message structure used for communicating with the config server
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
	// OPGet is the operation for getting the config. Refer to README.md for the protocol
	OPGet = "get"

	// OPConnect is the connect operation
	OPConnect = "connect"

	// OPPing is the ping operation
	OPPing = "ping"

	// OPPong is the ping response
	OPPong = "pong"

	// OPPeersChanged is the peers_changed push operation
	OPPeersChanged = "peers_changed"

	// OPConfigChanged is the config changed push operation
	OPConfigChanged = "config_changed"

	TypeRequest  = "request"  // message type request
	TypeResponse = "response" // message type response
	TypePush     = "push"     // message type push
)

// ToBytes converts the message to byte array for sending to socket
func (m *Message) ToBytes() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}
