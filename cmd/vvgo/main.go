//go:generate go run github.com/virtual-vgo/vvgo/tools/version

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/facebook"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"os"
)

var logger = log.Logger()

type Config struct {
	ApiConfig      api.ServerConfig `envconfig:"api"`
	RedisConfig    redis.Config     `envconfig:"redis"`
	MinioConfig    storage.Config   `envconfig:"minio"`
	DiscordConfig  discord.Config   `envconfig:"discord"`
	FacebookConfig facebook.Config  `envconfig:"facebook"`
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

	storage.Initialize(config.MinioConfig)
	redis.Initialize(config.RedisConfig)
	discord.Initialize(config.DiscordConfig)
	facebook.Initialize(config.FacebookConfig)

	apiServer := api.NewServer(ctx, config.ApiConfig)
	if err := apiServer.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
	}
}
