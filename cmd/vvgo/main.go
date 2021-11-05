package main

import (
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/server"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	var showVersion bool
	var showConfig bool
	var envFile string
	flag.BoolVar(&showVersion, "version", false, "show version and quit")
	flag.BoolVar(&showConfig, "env-usage", false, "show environment variable configuration")
	flag.StringVar(&envFile, "env-file", "", "file with environment variables")
	flag.Parse()

	switch {
	case showVersion:
		fmt.Println(version.String())
		os.Exit(0)
	case showConfig:
		_ = envconfig.Usage("", &config.Config)
		os.Exit(0)
	case envFile != "":
		config.ProcessEnvFile(envFile)
	default:
		config.ProcessEnv()
	}

	apiServer := server.NewServer(config.Config.VVGO.ListenAddress)
	logger.Println("listening on " + config.Config.VVGO.ListenAddress)

	go func() {
		if err := apiServer.ListenAndServe(); err != nil {
			logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	logger.Printf("caught %s", <-sigCh)
	if err := apiServer.Close(); err != nil {
		logger.WithError(err).Fatal("apiServer.Close() failed")
	}
	logger.Println("shutdown complete")
	os.Exit(0)
}
