//go:generate go run github.com/virtual-vgo/vvgo/tools/version

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"os"
)

var logger = log.Logger()

type Config struct {
	Api     api.ServerConfig
	Redis   redis.Config
	Minio   storage.Config
	Discord discord.Config
}

func (x *Config) ParseFile(name string) {
	logger := logger.WithField("file_name", name)
	file, err := os.Open(name)
	if err != nil {
		logger.WithError(err).Fatal("failed to open config")
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(x); err != nil {
		logger.WithError(err).Fatal("failed to parse config")
	}
}

type Flags struct {
	ShowVersion bool
	ShowConfig  bool
	ConfigPath  string
}

func (x *Flags) Parse() {
	flag.BoolVar(&x.ShowConfig, "print-config", false, "print config and quit")
	flag.BoolVar(&x.ShowVersion, "version", false, "show version and quit")
	flag.StringVar(&x.ConfigPath, "conf", "/etc/vvgo/config.json", "config file")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	var flags Flags
	flags.Parse()
	var config Config
	config.ParseFile(flags.ConfigPath)

	switch {
	case flags.ShowVersion:
		fmt.Println(version.String())
		os.Exit(0)
	case flags.ShowConfig:
		configJSON, _ := json.MarshalIndent(config, "", "  ")
		fmt.Println(string(configJSON))
		os.Exit(0)
	}

	storage.Initialize(config.Minio)
	redis.Initialize(config.Redis)
	discord.Initialize(config.Discord)
	apiServer := api.NewServer(ctx, config.Api)
	if err := apiServer.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
	}
}
