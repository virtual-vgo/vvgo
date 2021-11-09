package redis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mediocregopher/radix/v3"
	log "github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/errors"
	"github.com/virtual-vgo/vvgo/pkg/models/apilog"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"time"
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

var Client struct{ *radix.Pool }

func init() {
	radixPool, err := radix.NewPool(config.Config.Redis.Network, config.Config.Redis.Address, config.Config.Redis.PoolSize)
	if err != nil {
		log.WithError(err).Fatalf("radix.NewPool() failed")
	}
	Client.Pool = radixPool
}

func ReadSheet(ctx context.Context, spreadsheetName string, name string) ([][]interface{}, error) {
	var buf bytes.Buffer
	key := "sheets:" + spreadsheetName + ":" + name
	if err := Do(ctx, Cmd(&buf, GET, key)); err != nil {
		return nil, errors.RedisFailure(err)
	}

	if buf.Len() == 0 {
		return nil, nil
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
	if err := Do(ctx, Cmd(nil, SET, key, buf.String())); err != nil {
		return errors.RedisFailure(err)
	}
	return nil
}

func WriteLog(ctx context.Context, entry apilog.Entry) error {
	timestamp := fmt.Sprintf("%f", time.Duration(entry.StartTime.UnixNano()).Seconds())
	entry.Version = version.Get()
	var data bytes.Buffer
	if err := json.NewEncoder(&data).Encode(entry); err != nil {
		log.WithError(err).Error("json.Encode() failed")
	}
	return Do(ctx, Cmd(nil, ZADD, apilog.RedisKey, timestamp, data.String()))
}

const ZRANGEBYSCORE = "ZRANGEBYSCORE"
const GET = "GET"
const SET = "SET"
const ZADD = "ZADD"

func Do(_ context.Context, a Action) error {
	start := time.Now()
	if err := Client.Pool.Do(radix.Cmd(a.Rcv, a.Cmd, a.Args...)); err != nil {
		return err
	}

	var argBytes int
	for _, arg := range a.Args {
		argBytes += len(arg)
	}
	entry := apilog.RedisQuery{
		StartTime:       start,
		Cmd:             a.Cmd,
		ArgLen:          len(a.Args),
		ArgBytes:        argBytes,
		DurationSeconds: time.Since(start).Seconds(),
	}
	log.WithFields(entry.Fields()).Info("redis client: query completed")
	return nil
}
