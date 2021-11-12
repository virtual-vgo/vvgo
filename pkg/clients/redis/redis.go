package redis

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/mediocregopher/radix/v3"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/api/mixtape"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/errors"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"strconv"
	"strings"
)

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

var Client struct{ *radix.Pool }

func init() {
	radixPool, err := radix.NewPool(config.Env.Redis.Network, config.Env.Redis.Address, config.Env.Redis.PoolSize)
	if err != nil {
		logrus.WithError(err).Fatalf("radix.NewPool() failed")
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
	span, spanOk := tracing.NewSpanFromContext(ctx, "redis query")
	metrics, err := DoUntraced(a)

	if spanOk {
		tracing.WriteSpan(span.WithRedisQuery(metrics).WithError(err))
	} else {
		logger.Warn("redis client: invalid trace context")
	}
	return err
}

func DoUntraced(a Action) (tracing.RedisQueryMetrics, error) {
	metrics := tracing.NewRedisQueryMetrics(a.Cmd, a.Args)
	err := Client.Pool.Do(radix.Cmd(a.Rcv, a.Cmd, a.Args...))
	switch err {
	case nil:
		logger.
			WithFields(metrics.Fields()).
			WithError(err).
			Warn("redis client: query completed with error")
	default:
		logger.
			WithFields(metrics.Fields()).
			Info("redis client: query completed")
	}
	return metrics, err
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
