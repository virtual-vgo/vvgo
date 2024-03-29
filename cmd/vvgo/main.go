package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/clients/cloudflare"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/server"
	"github.com/virtual-vgo/vvgo/pkg/server/cron/which_time"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()

	rand.New(rand.NewSource(time.Now().UnixNano()))
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	var showVersion bool
	var showEnvUsage bool
	var showRuntimeConfig bool
	var envFile string
	flag.BoolVar(&showVersion, "version", false, "show version and quit")
	flag.BoolVar(&showEnvUsage, "env-usage", false, "show environment variable configuration")
	flag.BoolVar(&showRuntimeConfig, "runtime-config", false, "show runtime configuration")
	flag.StringVar(&envFile, "env-file", "", "file with environment variables")
	flag.Parse()

	switch {
	case showVersion:
		fmt.Println(version.String())
		os.Exit(0)
	case showEnvUsage:
		_ = envconfig.Usage("", &config.Config)
		os.Exit(0)
	case envFile != "":
		config.ProcessEnvFile(envFile)
	default:
		config.ProcessEnv()
	}

	if showRuntimeConfig {
		configJSON, _ := json.MarshalIndent(config.Config, "", "  ")
		fmt.Println(string(configJSON))
		os.Exit(0)
	}

	apiServer := server.NewServer(config.Config.VVGO.ListenAddress)
	logger.Println("http server: listening on " + config.Config.VVGO.ListenAddress)

	if !config.Config.Development {
		go cloudflare.PurgeCache()

		_, err := discord.CreateMessage(ctx, discord.VVGOChannelWebDevelopers, discord.CreateMessageParams{
			Embed: &discord.Embed{
				Title:       "🍏 Fresh VVGO Deployment",
				Description: fmt.Sprintf("**Build Time:** %s\n**Git Sha:** `%s`", version.BuildTime(), version.Get().GitSha),
			},
		})
		if err != nil {
			logger.HttpDoFailure(ctx, err)
		}
	}

	go func() {
		channelId := discord.VVGOChannelJacksonsSandbox
		if !config.Config.Development {
			channelId = discord.VVGOChannelTimezones
		}

		which_time.WhichTime(ctx, channelId)
		for range time.Tick(30 * time.Second) {
			which_time.WhichTime(ctx, channelId)
		}
	}()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		logger.Printf("http server: caught %s", <-sigCh)
		if err := apiServer.Close(); err != nil {
			logger.WithError(err).Fatal("apiServer.Close() failed")
		}
	}()

	if err := apiServer.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
		}
	}
	logger.Println("http server: closed")
	os.Exit(0)
}
