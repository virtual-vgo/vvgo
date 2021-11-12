package main

import (
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	http2 "github.com/virtual-vgo/vvgo/pkg/api/http"
	"github.com/virtual-vgo/vvgo/pkg/api/version"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
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
		_ = envconfig.Usage("", &config.Env)
		os.Exit(0)
	case envFile != "":
		config.ProcessEnvFile(envFile)
	default:
		config.ProcessEnv()
	}

	apiServer := http2.NewServer(config.Env.VVGO.ListenAddress)
	logger.Println("http server: listening on " + config.Env.VVGO.ListenAddress)

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
