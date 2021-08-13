package redis

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/mediocregopher/radix/v3"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"strings"
)

var logger = log.New()
var client *Client

type Client struct {
	config Config
	pool   *radix.Pool
}

type Config struct {
	Network  string `default:"tcp"`
	Address  string `default:"localhost:6379"`
	PoolSize int    `default:"10"`
}

func Initialize(config Config) {
	client = NewClientMust(config)
}

func InitializeFromEnv() {
	var config Config
	envconfig.MustProcess("REDIS", &config)
	Initialize(config)
}

func Do(ctx context.Context, a Action) error {
	return client.Do(ctx, a)
}

func NewClient(config Config) (*Client, error) {
	if config.Network == "" {
		config.Network = "tcp"
	}
	if config.Address == "" {
		config.Address = "localhost:6379"
	}
	if config.PoolSize == 0 {
		config.PoolSize = 10
	}
	radixPool, err := radix.NewPool(config.Network, config.Address, config.PoolSize)
	if err != nil {
		return nil, err
	}
	return &Client{pool: radixPool, config: config}, nil
}

func NewClientMust(config Config) *Client {
	client, err := NewClient(config)
	if err != nil {
		logger.WithError(err).Fatal("redis.NewClient() failed")
	}
	return client
}

type Action struct {
	cmd         string
	args        []string
	radixAction radix.Action
}

func Cmd(rcv interface{}, cmd string, args ...string) Action {
	return Action{
		cmd:         cmd,
		args:        args,
		radixAction: radix.Cmd(rcv, cmd, args...),
	}
}

func (x *Client) Do(_ context.Context, a Action) error {
	logger.Infof("redis query: %s %s", a.cmd, strings.Join(a.args, " "))
	return x.pool.Do(a.radixAction)
}
