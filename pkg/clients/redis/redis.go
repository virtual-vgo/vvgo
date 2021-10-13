package redis

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/mediocregopher/radix/v3"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/errors"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"strings"
)

type Action struct {
	Rcv  interface{}
	Cmd  string
	Args []string
}

func Cmd(rcv interface{}, cmd string, args ...string) Action {
	return Action{
		Rcv:  rcv,
		Cmd:  cmd,
		Args: args,
	}
}

type Client struct{ pool *radix.Pool }

var client = NewClientMust()

func NewClientMust() *Client {
	radixPool, err := radix.NewPool(config.Config.Redis.Network, config.Config.Redis.Address, config.Config.Redis.PoolSize)
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

func ReadSheet(ctx context.Context, spreadsheetName string, name string) ([][]interface{}, error) {
	var buf bytes.Buffer
	key := "sheets:" + spreadsheetName + ":" + name
	if err := Do(ctx, Cmd(&buf, "GET", key)); err != nil {
		return nil, errors.RedisFailure(err)
	}

	var values [][]interface{}
	if err := json.NewDecoder(&buf).Decode(&values); err != nil {
		return nil, errors.JsonDecodeFailure(err)
	}
	return values, nil
}

func WriteSheet(ctx context.Context, spreadsheetName, name string, values [][]interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&values); err != nil {
		return errors.JsonEncodeFailure(err)
	}

	key := "sheets:" + spreadsheetName + ":" + name
	if err := Do(ctx, Cmd(nil, "SET", key, buf.String())); err != nil {
		return errors.RedisFailure(err)
	}
	return nil
}

func Do(_ context.Context, a Action) error {
	truncArgs := func(args []string) string {
		argString := strings.Join(args, " ")
		if len(args) > 30 {
			argString = argString[:30] + "..."
		}
		return argString
	}

	logger.WithField("cmd", a.Cmd).Infof("redis query: %s %s", a.Cmd, truncArgs(a.Args))
	return client.pool.Do(radix.Cmd(a.Rcv, a.Cmd, a.Args...))
}
