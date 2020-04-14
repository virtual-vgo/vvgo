//go:generate go run github.com/virtual-vgo/vvgo/tools/version

package main

import (
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"os"
	"sync"
	"time"
)

var logger = log.Logger()

type Config struct {
	InitializeStorage bool             `split_words:"true" default:"false"`
	StorageConfig     storage.Config   `envconfig:"storage"`
	ApiConfig         api.ServerConfig `envconfig:"api"`
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
		fmt.Println(version.String())
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
	var config Config
	config.ParseEnv()
	config.ParseFlags()

	storage := storage.NewClient(config.StorageConfig)
	if storage == nil {
		os.Exit(1)
	}

	database := api.NewStorage(storage, config.ApiConfig)
	if database == nil {
		os.Exit(1)
	}

	if config.InitializeStorage {
		initializeStorage(database)
	}

	apiServer := api.NewServer(config.ApiConfig, database)
	if apiServer == nil {
		os.Exit(1)
	}
	logger.Fatal(apiServer.ListenAndServe())
}

func initializeStorage(db *api.Storage) {
	var wg sync.WaitGroup
	for _, initFunc := range []func() bool{
		db.Parts.Init,
	} {
		wg.Add(1)
		go func(initFunc func() bool) {
			defer wg.Done()
			timeout := time.NewTicker(5 * time.Second)
			retryInterval := time.NewTicker(500 * time.Millisecond)
			defer retryInterval.Stop()
			defer timeout.Stop()
			for range retryInterval.C {
				if ok := initFunc(); ok {
					return
				}
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
