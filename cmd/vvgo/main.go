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
	InitializeStorage bool             `envconfig:"initialize_storage"`
	StorageConfig     storage.Config   `envconfig:"storage"`
	ApiConfig         api.ServerConfig `envconfig:"api"`
}

func NewDefaultConfig() Config {
	return Config{
		InitializeStorage: false,
		ApiConfig: api.ServerConfig{
			ListenAddress:       ":8080",
			MaxContentLength:    1e6,
			MemberBasicAuthUser: "admin",
			MemberBasicAuthPass: "admin",
			SheetsBucketName:    "sheets",
			ClixBucketName:      "clix",
			PartsBucketName:     "parts",
			PartsLockerKey:      "parts.lock",
			AdminToken:          "admin",
			PrepRepToken:        "prep-rep",
		},
		StorageConfig: storage.Config{
			Minio: storage.MinioConfig{
				Endpoint:  "localhost:9000",
				Region:    "sfo2",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				UseSSL:    false,
			},
			Redis: storage.RedisConfig{
				Address: "localhost:6379",
			},
		},
	}
}

func (x *Config) ParseEnv() {
	err := envconfig.Process("", x)
	if err != nil {
		logger.Fatal(err)
	}
}

func (x Config) ParseFlags() {
	var showReleaseTags bool
	flag.BoolVar(&showReleaseTags, "release-tags", false, "show release tags and quit")
	flag.Parse()
	if showReleaseTags {
		for _, tag := range version.ReleaseTags() {
			fmt.Fprintf(os.Stdout, "%s\n", tag)
		}
		os.Exit(0)
	}
}

func main() {
	config := NewDefaultConfig()
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
