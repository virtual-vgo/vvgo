package redis

import (
	"context"
	"github.com/mediocregopher/radix/v3"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
)

type Client struct {
	pool *radix.Pool
}

type Config struct {
	Network  string `default:"tcp"`
	Address  string `default:"localhost:6379"`
	PoolSize int    `split_words:"true" default:"10"`
}

func NewClient(config Config) (*Client, error) {
	radixPool, err := radix.NewPool(config.Network, config.Address, config.PoolSize)
	if err != nil {
		return nil, err
	}
	return &Client{pool: radixPool}, nil
}

type Action struct {
	cmd         string
	radixAction radix.Action
}

func Cmd(rcv interface{}, cmd string, args ...string) Action {
	return Action{
		cmd:         cmd,
		radixAction: radix.Cmd(rcv, cmd, args...),
	}
}

func (x *Client) Do(ctx context.Context, a Action) error {
	_, span := tracing.StartSpan(ctx, "redis.Client.Do()")
	span.AddField("command", a.cmd)
	defer span.Send()
	return x.pool.Do(a.radixAction)
}
