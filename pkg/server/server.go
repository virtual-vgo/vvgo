package server

import (
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
)

type Server struct {
	*http.Server
}

func NewServer(listenAddress string) *Server {
	return &Server{
		Server: &http.Server{
			Addr:     listenAddress,
			Handler:  http_wrappers.Handler(Routes()),
			ErrorLog: logger.StdLogger(),
		},
	}
}
