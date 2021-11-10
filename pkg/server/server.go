package server

import (
	logurs "github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"log"
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
			ErrorLog: log.New(logurs.New().Writer(), "", 0),
		},
	}
}
