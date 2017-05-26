package cfgsrv

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tonjun/gostore"
	"github.com/tonjun/pubsub"
)

type ConnectHandler struct {
	store  gostore.Store
	config *map[string]interface{}
	opts   *Options

	reqID    int64
	reqIDMtx sync.Mutex
}

func NewConnectHandler(store gostore.Store, cfg *map[string]interface{}, opts *Options) Handler {
	h := &ConnectHandler{
		store:  store,
		config: cfg,
		opts:   opts,
	}
	h.store.OnListDidChange(h.onListDidChange)
	return h
}

func (h *ConnectHandler) ProcessMessage(m *Message, c pubsub.Conn) {

	if m.OP != OPConnect {
		return
	}

	peers := make([]string, 0)

	items, _, err := h.store.ListGet("peers")
	//log.Printf("items: %v", items)
	if err != nil {
		log.Printf("ListGet ERROR: %s", err.Error())
		return
	}
	if len(items) > 0 {

		// broadcast
		for _, item := range items {
			peers = append(peers, item.Value.(string))
		}

		peers = append(peers, m.Addr)

	} else {
		peers = append(peers, m.Addr)
	}

	h.store.Put(&gostore.Item{
		ID:    m.Addr,
		Key:   m.Addr,
		Value: c,
	}, 0)

	h.store.Put(&gostore.Item{
		ID:    fmt.Sprintf("%d", c.ID()),
		Key:   fmt.Sprintf("%d", c.ID()),
		Value: m.Addr,
	}, 0)

	h.store.Put(&gostore.Item{
		ID:    fmt.Sprintf("%s-ping", m.Addr),
		Key:   fmt.Sprintf("%s-ping", m.Addr),
		Value: m.Addr,
	}, time.Duration(h.opts.Timeout)*time.Second)

	resp := &Message{
		OP:     OPConnect,
		Type:   TypeResponse,
		ID:     m.ID,
		Config: h.config,
		Peers:  peers,
	}

	c.Send(resp.ToBytes())

	h.store.ListPush("peers", &gostore.Item{
		ID:    m.Addr,
		Key:   "peers",
		Value: m.Addr,
	})

}

func (h *ConnectHandler) Close() {
}

func (h *ConnectHandler) genReqID() string {
	h.reqIDMtx.Lock()
	defer h.reqIDMtx.Unlock()
	h.reqID = (h.reqID + 1) % 999999
	return fmt.Sprintf("%d", h.reqID)
}

func (h *ConnectHandler) pushPeers() {
	items, found, _ := h.store.ListGet("peers")
	if !found {
		return
	}
	mesg := &Message{
		OP:    OPPeersChanged,
		ID:    h.genReqID(),
		Type:  TypePush,
		Peers: make([]string, 0),
	}
	for _, item := range items {
		addr := item.Value.(string)
		mesg.Peers = append(mesg.Peers, addr)
	}

	// get the pubsub.Conn for each address and send the message
	for _, peer := range mesg.Peers {
		item, found, _ := h.store.Get(peer)
		if found {
			conn := item.Value.(pubsub.Conn)
			conn.Send(mesg.ToBytes())
		}
	}
}

func (h *ConnectHandler) onListDidChange(key string, items []*gostore.Item) {
	log.Printf("onListDidChange: key: \"%s\" len: %d", key, len(items))
	h.pushPeers()
}
