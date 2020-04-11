//go:generate go run github.com/virtual-vgo/vvgo/tools/version

package main

import (
	"flag"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"os"
	"strconv"
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
		},
		StorageConfig: storage.Config{
			MinioConfig: storage.MinioConfig{
				Endpoint:  "http://localhost:9000",
				Region:    "sfo2",
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				UseSSL:    false,
			},
			RedisConfig: storage.RedisConfig{
				Address: "http://localhost:6379",
			},
		},
	}
}

func (x *Config) ParseEnv() {
	x.InitializeStorage, _ = strconv.ParseBool(os.Getenv("INITIALIZE_STORAGE"))

	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		x.StorageConfig.MinioConfig.Endpoint = endpoint
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

	sheetsBucket := storage.NewBucket(api.SheetsBucketName)
	sheetsLocker := storage.NewLocker(api.SheetsLockerKey)
	if sheetsBucket == nil || sheetsLocker == nil {
		os.Exit(1)
	}
	sheets := sheets.Sheets{
		Bucket: sheetsBucket,
		Locker: sheetsLocker,
	}

	if config.InitializeStorage {
		if ok := initializeStorage(sheets); !ok {
			return
		}
		logger.Info("storage initialized")
	}

	apiServer := api.NewServer(config.ApiConfig, sheets)
	if apiServer == nil {
		os.Exit(1)
	}
	logger.Fatal(apiServer.ListenAndServe())
}

func initializeStorage(sheets sheets.Sheets) bool {
	retryInterval := time.NewTicker(500 * time.Millisecond)
	defer retryInterval.Stop()
	timeout := time.NewTicker(5 * time.Second)
	defer timeout.Stop()
	for range retryInterval.C {
		if ok := sheets.Init(); ok {
			return true
		}
		select {
		case <-timeout.C:
			return false
		default:
		}
	}
	return false
}
