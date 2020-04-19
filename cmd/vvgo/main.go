//go:generate go run github.com/virtual-vgo/vvgo/tools/version

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"os"
	"sync"
	"time"
)

var logger = log.Logger()

type Config struct {
	InitializeStorage bool             `split_words:"true" default:"false"`
	ApiConfig         api.ServerConfig `envconfig:"api"`
	TracingConfig     tracing.Config   `envconfig:"tracing"`
}

func (x *Config) ParseEnv() {
	err := envconfig.Process("", x)
	if err != nil {
		logger.Fatal(err)
	}
}

func (x Config) ParseFlags() {
	var showVersion, showReleaseTags, showEnvConfig bool
	flag.BoolVar(&showReleaseTags, "release-tags", false, "show release tags and quit")
	flag.BoolVar(&showVersion, "version", false, "show version and quit")
	flag.BoolVar(&showEnvConfig, "env-config", false, "show environment config and quit")
	flag.Parse()

	switch {
	case showVersion:
		fmt.Println(string(version.JSON()))
		os.Exit(0)
	case showReleaseTags:
		for _, tag := range version.ReleaseTags() {
			fmt.Println(tag)
		}
		os.Exit(0)
	case showEnvConfig:
		envconfig.Usage("", &x)
		os.Exit(0)
	}
}

func main() {
	ctx := context.Background()
	var config Config
	config.ParseEnv()
	config.ParseFlags()

	tracing.Initialize(config.TracingConfig)
	defer tracing.Close()

	database := api.NewStorage(ctx, config.ApiConfig)
	if database == nil {
		os.Exit(1)
	}

	if config.InitializeStorage {
		initializeStorage(ctx, database.Init)
	}

	apiServer := api.NewServer(config.ApiConfig, database)
	if err := apiServer.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
	}
}

func initializeStorage(ctx context.Context, initFuncs ...func(ctx context.Context) error) {
	var wg sync.WaitGroup
	for _, initFunc := range initFuncs {
		wg.Add(1)
		go func(initFunc func(ctx context.Context) error) {
			defer wg.Done()
			timeout := time.NewTicker(5 * time.Second)
			retryInterval := time.NewTicker(500 * time.Millisecond)
			defer retryInterval.Stop()
			defer timeout.Stop()
			for range retryInterval.C {
				err := initFunc(ctx)
				if err == nil {
					return
				}
				logger.WithError(err).Fatal("init() failed")
				select {
				case <-timeout.C:
					logger.Fatalf("failed to initialize storage")
				default:
				}
			}
		}(initFunc)
	}
	wg.Wait()
	logger.Info("storage initialized")
}
