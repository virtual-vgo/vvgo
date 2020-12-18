package redis

import (
	"context"
	"github.com/mediocregopher/radix/v3"
	"github.com/virtual-vgo/vvgo/pkg/log"
)

type Client struct {
	config Config
	pool   *radix.Pool
}

type Config struct {
	Network  string // Defaults to tcp.
	Address  string // Defaults to localhost:6379.
	PoolSize int    `split_words:"true" default:"10"`
}

var client *Client

func Initialize(config Config) {
	client = NewClientMust(config)
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
	radixPool, err := radix.NewPool(config.Network, config.Address, config.PoolSize)
	if err != nil {
		return nil, err
	}
	return &Client{pool: radixPool, config: config}, nil
}

var logger = log.Logger()

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
	return x.pool.Do(a.radixAction)
}
