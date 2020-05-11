package redis

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/mediocregopher/radix/v3"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
)

type Client struct {
	config Config
	pool   *radix.Pool
}

type Config struct {
	Network  string `default:"tcp"`
	Address  string `default:"localhost:6379"`
	PoolSize int    `split_words:"true" default:"10"`
}

var client *Client

func init() {
	var config Config
	envconfig.MustProcess("REDIS", &config)
	client = NewClientMust(config)
}

func Do(ctx context.Context, a Action) error {
	return client.Do(ctx, a)
}

func NewClient(config Config) (*Client, error) {
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

func (x *Client) Do(ctx context.Context, a Action) error {
	_, span := tracing.StartSpan(ctx, "redis.Client.Do()")
	span.AddField("command", a.cmd)
	span.AddField("args", a.args)
	defer span.Send()
	return x.pool.Do(a.radixAction)
}
