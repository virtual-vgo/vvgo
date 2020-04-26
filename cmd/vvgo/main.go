//go:generate go run github.com/virtual-vgo/vvgo/tools/version

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/virtual-vgo/vvgo/pkg/access"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"os"
)

var logger = log.Logger()

type Config struct {
	Secret            string            `envconfig:"vvgo_secret"`
	InitializeStorage bool              `split_words:"true" default:"false"`
	ApiConfig         api.ServerConfig  `envconfig:"api"`
	ApiStorageConfig  api.StorageConfig `envconfig:"api_storage"`
	TracingConfig     tracing.Config    `envconfig:"tracing"`
	StorageConfig     storage.Config    `envconfig:"storage"`
	SessionsConfig    access.Config     `envconfig:"sessions"`
	DiscordConfig     discord.Config    `envconfig:"discord"`
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

	var secret access.Secret
	sessionsStore := access.NewStore(secret, config.SessionsConfig)

	warehouse, err := storage.NewWarehouse(config.StorageConfig)
	if err != nil {
		logger.Fatal(err)
	}

	database := api.NewStorage(ctx, sessionsStore, warehouse, config.ApiStorageConfig)
	if database == nil {
		os.Exit(1)
	}

	if config.InitializeStorage {
		if err := database.Init(ctx); err != nil {
			logger.WithError(err).Fatal("failed to initialize storage")
		}
	}

	discordClient := discord.NewClient(config.DiscordConfig)
	apiServer := api.NewServer(config.ApiConfig, database, discordClient)
	if err := apiServer.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
	}
}
