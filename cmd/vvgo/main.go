package main

import (
	"flag"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"math/rand"
	"os"
	"time"
)

var logger = log.New()

type Flags struct {
	ShowVersion   bool
	ListenAddress string
}

func (x *Flags) Parse() {
	flag.BoolVar(&x.ShowVersion, "version", false, "show version and quit")
	flag.StringVar(&x.ListenAddress, "listen", "0.0.0.0:8080", "http listen address")
	flag.StringVar(&parse_config.FileName, "config-file", parse_config.FileName, "configuration file")
	flag.StringVar(&parse_config.Endpoint, "config-endpoint", parse_config.Endpoint, "endpoint for remote configuration")
	flag.StringVar(&parse_config.Session, "config-session", parse_config.Session, "session returned by https://vvgo.org/api/v1/session?with_roles=read_config")
	flag.Parse()
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	var flags Flags
	flags.Parse()

	switch {
	case flags.ShowVersion:
		fmt.Println(version.String())
		os.Exit(0)
	}

	redis.InitializeFromEnv()
	apiServer := api.NewServer(flags.ListenAddress)
	if err := apiServer.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
	}
}
