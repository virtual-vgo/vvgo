//go:generate go run github.com/virtual-vgo/vvgo/tools/version

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/honeycombio/beeline-go"
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
	InitializeStorage    bool             `split_words:"true" default:"false"`
	HoneycombWriteKey    string           `split_words:"true" default:""`
	HoneycombDataset     string           `split_words:"true" default:"development"`
	HoneycombServiceName string           `split_words:"true" default:"vvgo"`
	StorageConfig        storage.Config   `envconfig:"storage"`
	ApiConfig            api.ServerConfig `envconfig:"api"`
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

	initializeHoneycomb(config)
	defer beeline.Close()

	storage := storage.NewClient(config.StorageConfig)
	if storage == nil {
		os.Exit(1)
	}

	database := api.NewStorage(storage, config.ApiConfig)
	if database == nil {
		os.Exit(1)
	}

	if config.InitializeStorage {
		initializeStorage(ctx, database)
	}

	apiServer := api.NewServer(config.ApiConfig, database)
	if apiServer == nil {
		os.Exit(1)
	}
	logger.Fatal(apiServer.ListenAndServe())
}

func initializeStorage(ctx context.Context, db *api.Storage) {
	var wg sync.WaitGroup
	for _, initFunc := range []func(ctx context.Context) bool{
		db.Parts.Init,
	} {
		wg.Add(1)
		go func(initFunc func(ctx context.Context) bool) {
			defer wg.Done()
			timeout := time.NewTicker(5 * time.Second)
			retryInterval := time.NewTicker(500 * time.Millisecond)
			defer retryInterval.Stop()
			defer timeout.Stop()
			for range retryInterval.C {
				if ok := initFunc(ctx); ok {
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

func initializeHoneycomb(config Config) {
	beeline.Init(beeline.Config{
		ServiceName: config.HoneycombServiceName,
		WriteKey:    config.HoneycombWriteKey,
		Dataset:     config.HoneycombDataset,
	})
}
