package redis

import (
	"context"
	"github.com/mediocregopher/radix/v3"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"strings"
)

var logger = log.New()

type Client struct{ pool *radix.Pool }

var client *Client

func init() { client = NewClientMust() }

func Do(ctx context.Context, a Action) error { return client.Do(ctx, a) }

func NewClientMust() *Client {
	radixPool, err := radix.NewPool(parse_config.Config.Redis.Network, parse_config.Config.Redis.Address, parse_config.Config.Redis.PoolSize)
	if err != nil {
		logger.WithError(err).Fatal("redis.NewClient() failed")
		return nil
	}

	client := &Client{pool: radixPool}
	if err != nil {
		logger.WithError(err).Fatal("redis.NewClient() failed")
		return nil
	}
	return client
}

func (x *Client) Do(_ context.Context, a Action) error {
	args := strings.Join(a.args, " ")
	if len(args) > 30 {
		args = args[:30] + "..."
	}
	logger.WithField("cmd", a.cmd).Infof("redis query: %s %s", a.cmd, args)
	return x.pool.Do(a.radixAction)
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
