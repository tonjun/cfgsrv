package cfgsrv

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tonjun/gostore"
	"github.com/tonjun/pubsub"
)

// PingHandler is a config server handler that handles pong operation from client
// by periodically sending ping to all the connected clients and handling the response
type PingHandler struct {
	store    gostore.Store
	opts     *Options
	done     chan bool
	reqID    int64
	reqIDMtx sync.Mutex
}

// NewPingHandler creates a new instance of PingHandler
func NewPingHandler(store gostore.Store, opts *Options) Handler {
	h := &PingHandler{
		store: store,
		opts:  opts,
		done:  make(chan bool),
	}
	log.Printf("NewPingHandler timeout: %d", opts.Timeout)
	go h.pingLoop()
	store.OnItemDidExpire(h.onItemDidExpire)
	return h
}

// ProcessMessage is the implementation of the Handler interface
func (h *PingHandler) ProcessMessage(m *Message, c pubsub.Conn) {

	if m.OP != OPPong { // handle only pong operations
		return
	}

	// get addr given connection ID and update the mem store
	item, found, _ := h.store.Get(fmt.Sprintf("%d", c.ID()))
	if found {
		addr := item.Value.(string)
		log.Printf("updating ping for addr: %s", addr)
		h.store.Put(&gostore.Item{
			ID:    fmt.Sprintf("%s-ping", addr),
			Key:   fmt.Sprintf("%s-ping", addr),
			Value: addr,
		}, time.Duration(h.opts.Timeout)*time.Second)
	}
}

// Close closes the PingHandler
func (h *PingHandler) Close() {
	h.done <- true
}

func (h *PingHandler) pingLoop() {
	defer log.Printf("pingLoop done")

	p := (h.opts.Timeout * 1000) * 5 / 10
	log.Printf("pingLoop every: %d ms", p)

	//p = 1000

	for {
		select {
		case <-h.done:
			return

		case <-time.After(time.Duration(p) * time.Millisecond):
			items, found, _ := h.store.ListGet("peers")
			if found {
				m := &Message{
					OP:   OPPing,
					Type: TypeRequest,
					ID:   h.genReqID(),
				}
				for _, item := range items {
					addr := item.Value.(string)
					item, found, _ := h.store.Get(addr)
					if found {
						conn := item.Value.(pubsub.Conn)
						conn.Send(m.ToBytes())
					}
				}
			}

		}
	}
}

func (h *PingHandler) onItemDidExpire(item *gostore.Item) {
	addr := item.Value.(string)
	log.Printf("connection: \"%s\" expired key: \"%s\"", addr, item.Key)

	h.store.ListDel("peers", &gostore.Item{
		ID:    addr,
		Key:   "peers",
		Value: addr,
	})
}

func (h *PingHandler) genReqID() string {
	h.reqIDMtx.Lock()
	defer h.reqIDMtx.Unlock()
	h.reqID = (h.reqID + 1) % 999999
	return fmt.Sprintf("ping-%d", h.reqID)
}
