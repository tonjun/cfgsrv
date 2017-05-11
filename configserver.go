package cfgsrv

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/tonjun/pubsub"
	"github.com/tonjun/pubsub/wsserver"
)

type ConfigServer struct {
	opts   *Options
	srv    *wsserver.WSServer
	config map[string]interface{}
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
	}
}

func (s *ConfigServer) Start() error {

	d, err := ioutil.ReadFile(s.opts.ConfigFile)
	if err != nil {
		log.Printf("Read config file error: %s", err.Error())
		return err
	}
	log.Printf("config file contents: \"%s\"", string(d))
	//s.config = d

	// parse the config file
	err = json.Unmarshal(d, &s.config)
	if err != nil {
		log.Printf("Unmarshal config error: %s", err.Error())
		return err
	}
	log.Printf("cfg: %v", s.config)

	s.srv.OnMessage(s.onMessage)
	s.srv.Run()
	return nil
}

func (s *ConfigServer) Stop() {
	s.srv.Stop()
}

func (s *ConfigServer) onMessage(data []byte, c pubsub.Conn) {
	log.Printf("onMessage: %s", string(data))

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
		b, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}
		err = c.Send(b)
		if err != nil {
			log.Printf("send error: %s", err.Error())
		}
	}

}
