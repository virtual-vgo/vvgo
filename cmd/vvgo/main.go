package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/server"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	var showVersion bool
	var showEnvUsage bool
	var showRuntimeConfig bool
	var envFile string
	flag.BoolVar(&showVersion, "version", false, "show version and quit")
	flag.BoolVar(&showEnvUsage, "env-usage", false, "show environment variable configuration")
	flag.BoolVar(&showRuntimeConfig, "runtime-config", false, "show runtime configuration")
	flag.StringVar(&envFile, "env-file", "", "file with environment variables")
	flag.Parse()

	switch {
	case showVersion:
		fmt.Println(version.String())
		os.Exit(0)
	case showEnvUsage:
		_ = envconfig.Usage("", &config.Config)
		os.Exit(0)
	case envFile != "":
		config.ProcessEnvFile(envFile)
	default:
		config.ProcessEnv()
	}

	if showRuntimeConfig {
		configJSON, _ := json.MarshalIndent(config.Config, "", "  ")
		fmt.Println(string(configJSON))
		os.Exit(0)
	}

	apiServer := server.NewServer(config.Config.VVGO.ListenAddress)
	logger.Println("http server: listening on " + config.Config.VVGO.ListenAddress)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		logger.Printf("http server: caught %s", <-sigCh)
		if err := apiServer.Close(); err != nil {
			logger.WithError(err).Fatal("apiServer.Close() failed")
		}
	}()

	if err := apiServer.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
		}
	}
	logger.Println("http server: closed")
	os.Exit(0)
}
