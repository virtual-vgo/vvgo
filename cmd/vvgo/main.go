//go:generate go run github.com/virtual-vgo/vvgo/tools/version

package main

import (
	"flag"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
	"os"
	"strconv"
)

var logger = log.Logger()

type Config struct {
	Minio storage.MinioConfig
	Api   api.Config
}

func NewDefaultConfig() Config {
	return Config{
		Minio: storage.MinioConfig{
			Endpoint:  "localhost:9000",
			Region:    "sfo2",
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
			UseSSL:    false,
		},
		Api: api.Config{
			MaxContentLength: 1e6,
			BasicAuthUser:    "admin",
			BasicAuthPass:    "admin",
		},
	}
}

func (x *Config) ParseEnv() {
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

	apiServer := api.NewServer(
		storage.NewMinioDriverMust(config.Minio),
		config.Api,
	)
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: apiServer,
	}
	logger.Fatal(httpServer.ListenAndServe())
}
