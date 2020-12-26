package main

import (
	"flag"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"math/rand"
	"os"
	"time"
)

var logger = log.Logger()

type Flags struct {
	ShowVersion   bool
	ListenAddress string
}

func (x *Flags) Parse() {
	flag.BoolVar(&x.ShowVersion, "version", false, "show version and quit")
	flag.StringVar(&x.ListenAddress, "listen", "0.0.0.0:8080", "http listen address")
	flag.Parse()
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	var flags Flags
	flags.Parse()

	switch {
	case flags.ShowVersion:
		fmt.Println(version.String())
		os.Exit(0)
	}

	redis.InitializeFromEnv()
	apiServer := api.NewServer(flags.ListenAddress)
	if err := apiServer.ListenAndServe(); err != nil {
		logger.WithError(err).Fatal("apiServer.ListenAndServe() failed")
	}
}
