package cfgsrv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/tonjun/gostore"
	"github.com/tonjun/pubsub"
	"github.com/tonjun/pubsub/wsserver"
)

type ConfigServer struct {
	opts   *Options
	srv    *wsserver.WSServer
	config map[string]interface{}
	store  gostore.Store

	reqID    int64
	reqIDMtx sync.Mutex
}

type Options struct {
	ListenAddr string
	ConfigFile string
	Timeout    int32
}

func NewConfigServer(opts *Options) *ConfigServer {
	return &ConfigServer{
		opts: opts,
		srv: wsserver.NewWSServer(&wsserver.Options{
			ListenAddr: opts.ListenAddr,
			Path:       "/",
		}),
		store: gostore.NewStore(),
	}
}

func (s *ConfigServer) Start() error {

	d, err := ioutil.ReadFile(s.opts.ConfigFile)
	if err != nil {
		log.Printf("Read config file error: %s", err.Error())
		return err
	}
	//log.Printf("config file contents: \"%s\"", string(d))
	//s.config = d

	// parse the config file
	err = json.Unmarshal(d, &s.config)
	if err != nil {
		log.Printf("Unmarshal config error: %s", err.Error())
		return err
	}
	//log.Printf("cfg: %v", s.config)

	s.store.Init()

	s.srv.OnMessage(s.onMessage)
	s.srv.OnConnectionWillClose(s.onConnectionWillClose)
	s.srv.Run()
	return nil
}

func (s *ConfigServer) Stop() {
	s.srv.Stop()
	s.store.Close()
}

func (s *ConfigServer) onMessage(data []byte, c pubsub.Conn) {
	//log.Printf("onMessage: %s", string(data))

	req := &Message{}
	err := json.Unmarshal(data, req)
	if err != nil {
		log.Printf("onMessage parse error: %s", err.Error())
		return
	}
	log.Printf("operation: %s", req.OP)

	if req.OP == OPGet {
		resp := &Message{
			OP:     OPGet,
			Type:   TypeResponse,
			ID:     req.ID,
			Config: s.config,
		}
		s.send(c, resp)

	} else if req.OP == OPConnect {

		log.Printf("Connect: addr: %s", req.Addr)

		peers := make([]string, 0)

		items, _, err := s.store.ListGet("peers")
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

			peers = append(peers, req.Addr)

		} else {
			peers = append(peers, req.Addr)
		}

		s.store.ListPush("peers", &gostore.Item{
			ID:    req.Addr,
			Key:   "peers",
			Value: req.Addr,
		})

		s.pushPeers()

		s.store.Put(&gostore.Item{
			ID:    req.Addr,
			Key:   req.Addr,
			Value: c,
		}, 0)

		s.store.Put(&gostore.Item{
			ID:    fmt.Sprintf("%d", c.ID()),
			Key:   fmt.Sprintf("%d", c.ID()),
			Value: req.Addr,
		}, 0)

		resp := &Message{
			OP:     OPConnect,
			Type:   TypeResponse,
			ID:     req.ID,
			Config: s.config,
			Peers:  peers,
		}
		s.send(c, resp)

	}

}

func (s *ConfigServer) send(c pubsub.Conn, m *Message) {
	b, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	err = c.Send(b)
	if err != nil {
		log.Printf("send error: %s", err.Error())
	}
}

func (s *ConfigServer) onConnectionWillClose(c pubsub.Conn) {
	item, found, _ := s.store.Get(fmt.Sprintf("%d", c.ID()))
	if found {
		addr := item.Value.(string)
		log.Printf("connectin closed for addr: %s", addr)
	} else {
		log.Printf("connection %d not found in store", c.ID())
	}
}

func (s *ConfigServer) pushPeers() {
	items, found, _ := s.store.ListGet("peers")
	if !found {
		return
	}
	mesg := &Message{
		OP:    OPPeersChanged,
		ID:    s.genReqID(),
		Type:  TypePush,
		Peers: make([]string, 0),
	}
	for _, item := range items {
		addr := item.Value.(string)
		mesg.Peers = append(mesg.Peers, addr)
	}

	// get the pubsub.Conn for each address and send the message
	for _, peer := range mesg.Peers {
		item, found, _ := s.store.Get(peer)
		if found {
			conn := item.Value.(pubsub.Conn)
			s.send(conn, mesg)
		}
	}
}

func (s *ConfigServer) genReqID() string {
	s.reqIDMtx.Lock()
	defer s.reqIDMtx.Unlock()
	id := (s.reqID + 1) % 999999
	return fmt.Sprintf("%d", id)
}
