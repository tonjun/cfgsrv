package cfgsrv

import (
	"github.com/tonjun/pubsub/wsserver"
)

type ConfigServer struct {
	srv *wsserver.WSServer
}

type Options struct {
	ListenAddr string
	ConfigFile string
	Timeout    int32
}

func NewConfigServer(opts *Options) *ConfigServer {
	return &ConfigServer{
		srv: wsserver.NewWSServer(&wsserver.Options{
			ListenAddr: opts.ListenAddr,
			Path:       "/",
		}),
	}
}

func (s *ConfigServer) Start() {
	s.srv.Run()
}

func (s *ConfigServer) Stop() {
}
