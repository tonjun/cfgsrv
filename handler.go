package cfgsrv

import (
	"github.com/tonjun/pubsub"
)

type Handler interface {
	ProcessMessage(m *Message, c pubsub.Conn)
}
