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

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "show version and quit")
	flag.StringVar(&parse_config.ListenAddress, "listen", parse_config.ListenAddress, "http listen address")
	flag.StringVar(&parse_config.FileName, "config-file", parse_config.FileName, "configuration file")
	flag.StringVar(&parse_config.ServerURL, "server-url", parse_config.ServerURL, "url of the server")
	flag.Parse()

	switch {
	case showVersion:
		fmt.Println(version.String())
		os.Exit(0)
	}

	redis.InitializeFromEnv()
	apiServer := api.NewServer(parse_config.ListenAddress)
	if err := apiServer.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
	}
}
