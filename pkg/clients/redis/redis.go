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
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models/traces"
	"strings"
	"time"
)

const ZREVRANGEBYSCORE = "ZREVRANGEBYSCORE"
const ZRANGEBYSCORE = "ZRANGEBYSCORE"
const GET = "GET"
const INCR = "INCR"
const SET = "SET"
const ZADD = "ZADD"

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

func Do(ctx context.Context, a Action) error {
	var err error
	metrics := traces.NewRedisQueryMetrics(a.Cmd, a.Args)
	span, ok := traces.NewSpanFromContext(ctx, "redis query")
	if !ok {
		logger.Warn("redis client: invalid trace context")
	} else {
		defer func() { WriteSpan(span.WithRedisQuery(metrics).WithError(err)) }()
	}

	if err = Client.Pool.Do(radix.Cmd(a.Rcv, a.Cmd, a.Args...)); err != nil {
		logger.
			WithFields(metrics.Fields()).
			WithError(err).
			Warn("redis client: query completed with error")
	} else {
		logger.
			WithFields(metrics.Fields()).
			Info("redis client: query completed")
	}
	return err
}

func NewTrace(ctx context.Context, name string) (*traces.Span, error) {
	var traceId int64
	err := Client.Pool.Do(radix.Cmd(&traceId, INCR, traces.NextTraceIdRedisKey))
	if err != nil {
		return nil, err
	}
	trace := traces.NewTrace(ctx, uint64(traceId), name)
	return &trace, nil
}

func WriteSpan(span traces.Span) {
	if span.Duration == 0 {
		span = span.Finish()
	}
	timestamp := fmt.Sprintf("%f", time.Duration(span.StartTime.UnixNano()).Seconds())
	var data bytes.Buffer
	if err := json.NewEncoder(&data).Encode(span); err != nil {
		log.WithError(err).Error("json.Encode() failed")
		return
	}
	if err := Client.Pool.Do(radix.Cmd(nil, ZADD, traces.SpansRedisKey, timestamp, data.String())); err != nil {
		logger.RedisFailure(context.Background(), err)
		return
	}
}

func ListSpans(ctx context.Context, start, end time.Time) ([]traces.Span, error) {
	startString := fmt.Sprintf("%f", time.Duration(start.UnixNano()).Seconds())
	endString := fmt.Sprintf("%f", time.Duration(end.UnixNano()).Seconds())

	cmd := ZRANGEBYSCORE
	if end.Before(start) {
		cmd = ZREVRANGEBYSCORE
	}

	var entriesJSON []string
	if err := Do(ctx, Cmd(&entriesJSON, cmd, traces.SpansRedisKey, startString, endString)); err != nil {
		return nil, err
	}
	spans := make([]traces.Span, 0, len(entriesJSON))
	for _, logJSON := range entriesJSON {
		var entry traces.Span
		if err := json.NewDecoder(strings.NewReader(logJSON)).Decode(&entry); err != nil {
			logger.WithError(err).Error("json.Decode() failed")
		}
		spans = append(spans, entry)
	}
	return spans, nil
}
