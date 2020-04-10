//go:generate go run github.com/virtual-vgo/vvgo/tools/version

package main

import (
	"flag"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/sheet"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
	"os"
	"strconv"
)

var logger = log.Logger()

type Config struct {
	InitializeStorage bool

	Minio storage.MinioConfig
	Redis storage.RedisConfig
	Api   api.Config
}

func NewDefaultConfig() Config {
	return Config{
		InitializeStorage: false,
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
		Api: api.Config{
			MaxContentLength: 1e6,
			BasicAuthUser:    "admin",
			BasicAuthPass:    "admin",
		},
	}
}

func (x *Config) ParseEnv() {
	if initializeStorage := os.Getenv("INITIALIZE_STORAGE"); initializeStorage != "" {
		x.InitializeStorage, _ = strconv.ParseBool(initializeStorage)
	}

	if endpoint := os.Getenv("MINIO_ENDPOINT"); endpoint != "" {
		x.Minio.Endpoint = endpoint
	}
	if id := os.Getenv("MINIO_ACCESS_KEY"); id != "" {
		x.Minio.AccessKey = id
	}
	if key := os.Getenv("MINIO_SECRET_KEY"); key != "" {
		x.Minio.SecretKey = key
	}
	x.Minio.UseSSL, _ = strconv.ParseBool(os.Getenv("MINIO_USE_SSL"))

	if maxContentLength, _ := strconv.ParseInt(os.Getenv("API_MAX_CONTENT_LENGTH"), 10, 64); maxContentLength != 0 {
		x.Api.MaxContentLength = maxContentLength
	}

	if user := os.Getenv("BASIC_AUTH_USER"); user != "" {
		x.Api.BasicAuthUser = user
	}
	if pass := os.Getenv("BASIC_AUTH_PASS"); pass != "" {
		x.Api.BasicAuthPass = pass
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

	minioDriver := storage.NewMinioDriverMust(config.Minio)
	redisLocker := storage.NewRedisLocker(config.Redis)

	if config.InitializeStorage {
		logger.Info("initializing storage...")
		sheetStorage := sheet.Storage{RedisLocker: redisLocker, MinioDriver: minioDriver}
		sheetStorage.Init()
	}

	apiServer := api.NewServer(
		storage.NewMinioDriverMust(config.Minio),
		storage.NewRedisLocker(config.Redis),
		config.Api,
	)
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: apiServer,
		ErrorLog: log.StdLogger(),
	}
	logger.Fatal(httpServer.ListenAndServe())
}
