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
	var envFile string
	var configSession string
	var configEndpoint string
	flag.BoolVar(&showVersion, "version", false, "show version and quit")
	flag.BoolVar(&showConfig, "env-usage", false, "show environment variable configuration")
	flag.StringVar(&envFile, "env-file", "", "file with environment variables")
	flag.StringVar(&configSession, "config-session", "", "remote configuration session key from https://vvgo.org/api/v1/session?with_roles=read_config")
	flag.StringVar(&configEndpoint, "config-endpoint", "https://vvgo.org", "remote configuration endpoint")
	flag.Parse()

	switch {
	case showVersion:
		fmt.Println(version.String())
		os.Exit(0)
	case showConfig:
		_ = envconfig.Usage("", &parse_config.Config)
		os.Exit(0)
	case envFile != "":
		parse_config.ProcessEnvFile(envFile)
	case configEndpoint != "" && configSession != "":
		parse_config.ProcessEndpoint(configEndpoint, configSession)
	default:
		parse_config.ProcessEnv()
	}

	apiServer := api.NewServer(parse_config.Config.VVGO.ListenAddress)
	if err := apiServer.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
	}
}
