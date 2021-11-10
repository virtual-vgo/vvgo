package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models/mixtape"
	"github.com/virtual-vgo/vvgo/pkg/server"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func migrateMixtapes() {
	ctx := context.Background()
	projects, err := redis.ListMixtapeProjects(context.Background())
	if err != nil {
		logger.Fatal("redis.Do() failed:", err)
	}

	reversed := make([]mixtape.Project, len(projects))
	for i := range projects {
		reversed[i] = projects[len(projects)-1-i]
	}

	for _, proj := range reversed {
		if proj.Id == 0 {
			var id uint64
			if err := redis.Do(ctx, redis.Cmd(&id, redis.INCR, mixtape.NextProjectIdRedisKey)); err != nil {
				logger.Fatal("redis.Do() failed:", err)
			}
			proj.Id = id
			var projJSON bytes.Buffer
			if err := json.NewEncoder(&projJSON).Encode(proj); err != nil {
				logger.Fatal("json.Encode() failed:", err)
			}
			if err := redis.Do(ctx, redis.Cmd(nil, redis.HSET,
				mixtape.ProjectsRedisKey, redis.ObjectId(proj.Id).String(), projJSON.String())); err != nil {
				logger.Fatal("redis.Do() failed:", err)
			}
			if err := redis.Do(ctx, redis.Cmd(nil, redis.HDEL,
				mixtape.ProjectsRedisKey, proj.Name)); err != nil {
				logger.Fatal("redis.Do() failed:", err)
			}
		}
	}
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})

	var showVersion bool
	var showConfig bool
	var envFile string
	flag.BoolVar(&showVersion, "version", false, "show version and quit")
	flag.BoolVar(&showConfig, "env-usage", false, "show environment variable configuration")
	flag.StringVar(&envFile, "env-file", "", "file with environment variables")
	flag.Parse()

	switch {
	case showVersion:
		fmt.Println(version.String())
		os.Exit(0)
	case showConfig:
		_ = envconfig.Usage("", &config.Config)
		os.Exit(0)
	case envFile != "":
		config.ProcessEnvFile(envFile)
	default:
		config.ProcessEnv()
	}

	migrateMixtapes()

	apiServer := server.NewServer(config.Config.VVGO.ListenAddress)
	logger.Println("http server: listening on " + config.Config.VVGO.ListenAddress)

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
