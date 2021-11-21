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
	var showUsage bool
	var showConfig bool
	var envFile string
	flag.BoolVar(&showVersion, "version", false, "show version and quit")
	flag.BoolVar(&showUsage, "env-usage", false, "show environment variable usage and quit")
	flag.BoolVar(&showConfig, "show-config", false, "show runtime configuration and quit")
	flag.StringVar(&envFile, "env-file", "", "environment file (optional)")
	flag.Parse()

	if showVersion {
		fmt.Println(version.String())
		os.Exit(0)
	}

	if showUsage {
		_ = envconfig.Usage(config.EnvPrefix, &config.Config)
		os.Exit(0)
	}

	config.ProcessEnvFile(envFile)
	if showConfig {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(config.Config)
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
