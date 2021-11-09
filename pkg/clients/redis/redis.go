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
	"github.com/virtual-vgo/vvgo/pkg/version"
	"strings"
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

const LogsApiKey = "logs:api"

func WriteLog(ctx context.Context, entry *log.Entry) error {
	timestamp := fmt.Sprintf("%f", time.Duration(entry.Time.UnixNano()).Seconds())
	entry.WithField("api_version", version.Get())
	var data bytes.Buffer
	if err := json.NewEncoder(&data).Encode(entry.Data); err != nil {
		log.WithError(err).Error("json.Encode() failed")
	}
	return Do(ctx, Cmd(nil, ZADD, LogsApiKey, timestamp, data.String()))
}

func ListLogs(ctx context.Context, start time.Time, end time.Time) ([]string, error) {
	startString := fmt.Sprintf("%f", time.Duration(start.UnixNano()).Seconds())
	endString := fmt.Sprintf("%f", time.Duration(end.UnixNano()).Seconds())
	var entriesJSON []string

	err := Do(ctx, Cmd(&entriesJSON, ZRANGEBYSCORE, LogsApiKey, startString, endString))
	return entriesJSON, err
}

const ZRANGEBYSCORE = "ZRANGEBYSCORE"
const GET = "GET"
const SET = "SET"
const ZADD = "ZADD"

func Do(_ context.Context, a Action) error {
	truncArgs := func(args []string) string {
		argString := strings.Join(args, " ")
		if len(argString) > 64 {
			argString = argString[:61] + "..."
		}
		return argString
	}

	log.WithField("cmd", a.Cmd).Debugf("redis query: %s %s", a.Cmd, truncArgs(a.Args))
	return Client.Pool.Do(radix.Cmd(a.Rcv, a.Cmd, a.Args...))
}
