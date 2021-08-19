package main

import (
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"math/rand"
	"os"
	"time"
)

var logger = log.New()

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	var showVersion bool
	var showConfig bool
	flag.BoolVar(&showVersion, "version", false, "show version and quit")
	flag.BoolVar(&showConfig, "env-usage", false, "show environment variable configuration")
	flag.Parse()

	switch {
	case showVersion:
		fmt.Println(version.String())
		os.Exit(0)
	case showConfig:
		_ = envconfig.Usage("", &parse_config.Config)
		os.Exit(0)
	}

	apiServer := api.NewServer(parse_config.Config.VVGO.ListenAddress)
	if err := apiServer.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
	}
}
