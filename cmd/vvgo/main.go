//go:generate go run github.com/virtual-vgo/vvgo/tools/version

package main

import (
	"flag"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"os"
	"strconv"
	"sync"
	"time"
)

var logger = log.Logger()

type Config struct {
	InitializeStorage bool
	StorageConfig     storage.Config
	ApiConfig         api.ServerConfig
}

func NewDefaultConfig() Config {
	return Config{
		InitializeStorage: false,
		ApiConfig: api.ServerConfig{
			ListenAddress:    ":8080",
			MaxContentLength: 1e6,
			BasicAuthUser:    "admin",
			BasicAuthPass:    "admin",
			SheetsBucketName: "sheets",
			ClixBucketName:   "clix",
			PartsBucketName:  "parts",
			PartsLockerKey:   "parts.lock",
		},
		StorageConfig: storage.Config{
			MinioConfig: storage.MinioConfig{
				Endpoint:  "localhost:9000",
				Region:    "sfo2",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				UseSSL:    false,
			},
			RedisConfig: storage.RedisConfig{
				Address: "localhost:6379",
			},
		},
	}
}

func (x *Config) ParseEnv() {
	x.InitializeStorage, _ = strconv.ParseBool(os.Getenv("INITIALIZE_STORAGE"))

	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		x.StorageConfig.MinioConfig.Endpoint = endpoint
	}
	if arg := os.Getenv("MINIO_REGION"); arg != "" {
		x.StorageConfig.MinioConfig.Region = arg
	}
	if id := os.Getenv("MINIO_ACCESS_KEY"); id != "" {
		x.StorageConfig.MinioConfig.AccessKey = id
	}
	if key := os.Getenv("MINIO_SECRET_KEY"); key != "" {
		x.StorageConfig.MinioConfig.SecretKey = key
	}
	x.StorageConfig.MinioConfig.UseSSL, _ = strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))

	if address := os.Getenv("REDIS_ADDRESS"); address != "" {
		x.StorageConfig.RedisConfig.Address = address
	}

	if maxContentLength, _ := strconv.ParseInt(os.Getenv("API_MAX_CONTENT_LENGTH"), 10, 64); maxContentLength != 0 {
		x.ApiConfig.MaxContentLength = maxContentLength
	}
	if listenAddress := os.Getenv("LISTEN_ADDRESS"); listenAddress != "" {
		x.ApiConfig.ListenAddress = listenAddress
	}
	if user := os.Getenv("BASIC_AUTH_USER"); user != "" {
		x.ApiConfig.BasicAuthUser = user
	}
	if pass := os.Getenv("BASIC_AUTH_PASS"); pass != "" {
		x.ApiConfig.BasicAuthPass = pass
	}
	if arg := os.Getenv("SHEETS_BUCKET_NAME"); arg != "" {
		x.ApiConfig.SheetsBucketName = arg
	}
	if arg := os.Getenv("CLIX_BUCKET_NAME"); arg != "" {
		x.ApiConfig.ClixBucketName = arg
	}
	if arg := os.Getenv("PARTS_BUCKET_NAME"); arg != "" {
		x.ApiConfig.PartsBucketName = arg
	}
	if arg := os.Getenv("PARTS_LOCKER_KEY"); arg != "" {
		x.ApiConfig.PartsLockerKey = arg
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
