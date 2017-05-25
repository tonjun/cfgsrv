package cfgsrv

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/tonjun/gostore"
	"github.com/tonjun/pubsub"
	"github.com/tonjun/pubsub/wsserver"
)

type ConfigServer struct {
	opts     *Options
	srv      *wsserver.WSServer
	config   map[string]interface{}
	store    gostore.Store
	handlers []Handler
	timeout  int32
}

// Options is the config server options used in NewConfigServer
type Options struct {
	ListenAddr string // Websocket listen address
	ConfigFile string // JSON config file
	Timeout    int32  // Ping timeout in seconds
}

// NewConfigServer creates a new instance of ConfigServer
func NewConfigServer(opts *Options) *ConfigServer {
	return &ConfigServer{
		opts: opts,
		srv: wsserver.NewWSServer(&wsserver.Options{
			ListenAddr: opts.ListenAddr,
			Path:       "/",
		}),
		store:    gostore.NewStore(),
		handlers: make([]Handler, 0),
		timeout:  opts.Timeout,
	}
}

// Start starts the Config server
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

	ch := NewConnectHandler(s.store, &s.config, s.opts)
	s.handlers = append(s.handlers, ch)

	ph := NewPingHandler(s.store, s.opts)
	s.handlers = append(s.handlers, ph)

	s.store.Init()

	s.srv.OnMessage(s.onMessage)
	s.srv.OnConnectionWillClose(s.onConnectionWillClose)
	s.srv.Run()
	return nil
}

// Stop stops the config server
func (s *ConfigServer) Stop() {
	s.srv.Stop()
	s.store.Close()
	for _, h := range s.handlers {
		h.Close()
	}
}

// GetStore returns the in-memory k/v store
func (s *ConfigServer) GetStore() gostore.Store {
	return s.store
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

	// pass to all the handlers
	for _, h := range s.handlers {
		h.ProcessMessage(req, c)
	}

	if req.OP == OPGet {
		resp := &Message{
			OP:     OPGet,
			Type:   TypeResponse,
			ID:     req.ID,
			Config: s.config,
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
		// remove connection from mem store
		addr := item.Value.(string)
		log.Printf("connectin closed for addr: %s", addr)
		s.store.Del(fmt.Sprintf("%d", c.ID()))
		s.store.Del(addr)
	} else {
		log.Printf("connection %d not found in store", c.ID())
	}
}
