package redis

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/mediocregopher/radix/v3"
	log "github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/errors"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models/mixtape"
	"github.com/virtual-vgo/vvgo/pkg/models/traces"
	"strconv"
	"strings"
	"sync"
	"time"
)

const Network = "tcp"

const NewClientRetryWaitTime = 1 * time.Second

const (
	GET              = "GET"
	HDEL             = "HDEL"
	HGET             = "HGET"
	HGETALL          = "HGETALL"
	HSET             = "HSET"
	INCR             = "INCR"
	SET              = "SET"
	ZADD             = "ZADD"
	ZRANGEBYSCORE    = "ZRANGEBYSCORE"
	ZREVRANGEBYSCORE = "ZREVRANGEBYSCORE"
)

type ObjectId uint64

func StringToObjectId(str string) uint64 { id, _ := strconv.ParseUint(str, 10, 64); return id }
func (id ObjectId) String() string       { return strconv.FormatUint(uint64(id), 10) }

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

var client *radix.Pool

func getClient() *radix.Pool {
	initClient()
	return client
}

var initClientOnce = sync.Once{}

func initClient() {
	initClientOnce.Do(func() {
		ticker := time.NewTicker(NewClientRetryWaitTime)
		defer ticker.Stop()
		for attempt := 0; ; attempt++ {
			var err error
			client, err = radix.NewPool(Network,
				config.Config.Redis.Address,
				config.Config.Redis.PoolSize,
				radix.PoolConnFunc(func(network, addr string) (radix.Conn, error) {
					var dialOpts []radix.DialOpt

					if config.Config.Redis.Pass != "" {
						dialOpts = append(dialOpts,
							radix.DialAuthUser(config.Config.Redis.User, config.Config.Redis.Pass),
							radix.DialSelectDB(config.Config.Redis.UseDB))

					}

					if config.Config.Redis.UseTLS {
						dialOpts = append(dialOpts, radix.DialUseTLS(&tls.Config{InsecureSkipVerify: true}))
					}
					return radix.Dial("tcp", config.Config.Redis.Address, dialOpts...)
				}))
			if err != nil {
				log.WithField("attempt", attempt).WithError(err).Warnf("radix.Dial() failed")
				log.WithField("attempt", attempt).Warnf("retry after %v", NewClientRetryWaitTime)
				<-ticker.C
				continue
			}
			break
		}
	})
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
	initClient()
	var err error
	metrics := traces.NewRedisQueryMetrics(a.Cmd, a.Args)
	span, ok := traces.NewSpanFromContext(ctx, "redis query")
	if !ok {
		logger.Warn("redis client: invalid trace context")
	} else {
		defer func() { WriteSpan(span.WithRedisQuery(metrics).WithError(err)) }()
	}

	if err = getClient().Do(radix.Cmd(a.Rcv, a.Cmd, a.Args...)); err != nil {
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
	initClient()
	var traceId int64
	err := getClient().Do(radix.Cmd(&traceId, INCR, traces.NextTraceIdRedisKey))
	if err != nil {
		return nil, err
	}
	trace := traces.NewTrace(ctx, uint64(traceId), name)
	return &trace, nil
}

func WriteSpan(span traces.Span) {
	initClient()
	if span.Duration == 0 {
		span = span.Finish()
	}
	timestamp := fmt.Sprintf("%f", time.Duration(span.StartTime.UnixNano()).Seconds())
	var data bytes.Buffer
	if err := json.NewEncoder(&data).Encode(span); err != nil {
		log.WithError(err).Error("json.Encode() failed")
		return
	}
	if err := getClient().Do(radix.Cmd(nil, ZADD, traces.SpansRedisKey, timestamp, data.String())); err != nil {
		logger.RedisFailure(context.Background(), err)
		return
	}
}

func ListSpans(ctx context.Context, start, end time.Time) ([]traces.Span, error) {
	initClient()
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

func ListMixtapeProjects(ctx context.Context) ([]mixtape.Project, error) {
	var projectsJSON []string
	if err := Do(ctx, Cmd(&projectsJSON, "HVALS", mixtape.ProjectsRedisKey)); err != nil {
		return nil, errors.RedisFailure(err)
	}

	projects := make([]mixtape.Project, 0, len(projectsJSON))
	for _, projectJSON := range projectsJSON {
		var project mixtape.Project
		if err := json.NewDecoder(strings.NewReader(projectJSON)).Decode(&project); err != nil {
			logger.JsonDecodeFailure(ctx, err)
			continue
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func SaveMixtapeProject(ctx context.Context, project mixtape.Project) error {
	if project.Id == 0 {
		return errors.New("project id cannot be zero")
	}

	var projectJSON bytes.Buffer
	if err := json.NewEncoder(&projectJSON).Encode(project); err != nil {
		return err
	}

	return Do(ctx, Cmd(nil, HSET,
		mixtape.ProjectsRedisKey, strconv.FormatUint(project.Id, 16), projectJSON.String()))
}

func DeleteMixtapeProject(ctx context.Context, id uint64) error {
	return Do(ctx, Cmd(nil, HDEL,
		mixtape.ProjectsRedisKey, strconv.FormatUint(id, 16)))
}
