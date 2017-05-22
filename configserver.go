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
		store:    gostore.NewStore(),
		handlers: make([]Handler, 0),
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

	ch := NewConnectHandler(s.store, &s.config)
	s.handlers = append(s.handlers, ch)

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
		addr := item.Value.(string)
		log.Printf("connectin closed for addr: %s", addr)
	} else {
		log.Printf("connection %d not found in store", c.ID())
	}
}
