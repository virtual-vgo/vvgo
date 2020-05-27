package parts

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"strconv"
	"strings"
	"sync"
	"time"
)

var logger = log.Logger()

var (
	ErrInvalidPartName = fmt.Errorf("invalid part name")
)

type RedisParts struct {
	namespace string
}

func NewParts(namespace string) *RedisParts {
	return &RedisParts{namespace: namespace}
}

func (x *RedisParts) DeleteAll(ctx context.Context) error {
	localCtx, span := tracing.StartSpan(ctx, "RedisParts.List()")
	defer span.Send()
	partKeys, err := x.readIndex(localCtx)
	if err != nil {
		return err
	}

	allKeys := make([]string, 0, 1+3*len(partKeys))
	allKeys = append(allKeys, x.namespace+":parts:index")
	for _, key := range partKeys {
		allKeys = append(allKeys, x.namespace+":parts:"+key+":sheets", x.namespace+":parts:"+key+":clix")
	}
	return redis.Do(localCtx, redis.Cmd(nil, "DEL", allKeys...))
}

// List returns a slice of all parts in the database.
// If the first request to redis fails, this function will return an error.
// Subsequent errors will only be logged.
func (x *RedisParts) List(ctx context.Context) ([]Part, error) {
	localCtx, span := tracing.StartSpan(ctx, "RedisParts.List()")
	defer span.Send()

	// Read the all the part _keys_ first.
	partKeys, err := x.readIndex(localCtx)
	if err != nil {
		return nil, err
	}

	// Now we can read each individual part data.
	// Radix will automatically batch/pipeline requests for us, so we can queue up a bunch of requests in goroutines.
	// https://pkg.go.dev/github.com/mediocregopher/radix/v3?tab=doc#hdr-Implicit_pipelining
	parts := make([]Part, len(partKeys))
	var wg sync.WaitGroup
	wg.Add(2 * len(partKeys))
	for i := range partKeys {
		parts[i].DecodeRedisKey(partKeys[i])
		go func(key string, dest *Part) {
			sheetsKey := x.namespace + ":parts:" + key + ":sheets"
			dest.Sheets = readLinks(ctx, sheetsKey)
			wg.Done()
		}(partKeys[i], &parts[i])

		go func(key string, dest *Part) {
			clixKey := x.namespace + ":parts:" + key + ":clix"
			dest.Clix = readLinks(ctx, clixKey)
			wg.Done()
		}(partKeys[i], &parts[i])
	}
	wg.Wait()
	return parts, nil
}

func (x *RedisParts) readIndex(localCtx context.Context) ([]string, error) {
	var partKeys []string
	if err := redis.Do(localCtx, redis.Cmd(&partKeys, "ZRANGE", x.namespace+":parts:index", "0", "-1")); err != nil {
		return nil, err
	}
	return partKeys, nil
}

func readLinks(ctx context.Context, key string) []Link {
	var raw []string
	if err := redis.Do(ctx, redis.Cmd(&raw, "ZREVRANGE", key, "0", "-1")); err != nil {
		logger.WithError(err).Error("ZREVRANGE")
	}
	links := make([]Link, len(raw))
	for i := range links {
		links[i].DecodeRedisString(raw[i])
	}
	return links
}

// Save a slice of parts to redis.
// This function returns an error if the parts are invalid.
func (x *RedisParts) Save(ctx context.Context, parts []Part) error {
	localCtx, span := tracing.StartSpan(ctx, "RedisParts.Save")
	defer span.Send()

	if len(parts) == 0 {
		return nil
	}

	for i := range parts {
		if err := parts[i].Validate(); err != nil {
			return err
		}
	}

	// Like with List(), we'll leverage pipelining and queue up lots of goroutines.
	var wg sync.WaitGroup

	// Update the index with the new parts.
	wg.Add(1)
	go func() {
		x.saveIndex(parts, localCtx)
		wg.Done()
	}()

	// Update all the links.
	wg.Add(2 * len(parts))
	for i := range parts {
		go func(src *Part) {
			sheetsKey := x.namespace + ":parts:" + src.RedisKey() + ":sheets"
			x.saveLinks(localCtx, sheetsKey, src.Sheets)
			wg.Done()
		}(&parts[i])

		go func(src *Part) {
			clixKey := x.namespace + ":parts:" + src.RedisKey() + ":clix"
			x.saveLinks(localCtx, clixKey, src.Clix)
			wg.Done()
		}(&parts[i])
	}
	wg.Wait()
	return nil
}

func (x *RedisParts) saveIndex(parts []Part, ctx context.Context) {
	args := make([]string, 0, 1+2*len(parts))
	args = append(args, x.namespace+":parts:index")
	for i := range parts {
		score := strconv.Itoa(parts[i].ZScore())
		member := parts[i].RedisKey()
		args = append(args, score, member)
	}

	if err := redis.Do(ctx, redis.Cmd(nil, "ZADD", args...)); err != nil {
		logger.WithError(err).WithField("args", args).Error("ZADD")
	}
}

func (x *RedisParts) saveLinks(ctx context.Context, key string, links []Link) {
	if len(links) == 0 {
		return
	}
	args := make([]string, 0, 1+2*len(links))
	args = append(args, key)
	for i := range links {
		score := strconv.Itoa(links[i].ZScore())
		member := links[i].EncodeRedisString()
		args = append(args, score, member)
	}
	if err := redis.Do(ctx, redis.Cmd(nil, "ZADD", args...)); err != nil {
		logger.WithError(err).WithField("args", args).Error("ZADD")
	}
}

type Part struct {
	ID
	Clix   Links `json:"click,omitempty"`
	Sheets Links `json:"sheet,omitempty"`
}

func (x *Part) RedisKey() string {
	return x.ID.Project + ":" + x.ID.Name
}

func (x *Part) DecodeRedisKey(str string) {
	got := strings.SplitN(str, ":", 2)
	x.ID.Project = got[0]
	x.ID.Name = got[1]
}

func (x *Part) ZScore() int {
	projectScore, _ := strconv.Atoi(x.ID.Project[:2])
	return projectScore
}

type Link struct {
	ObjectKey string    `json:"object_key"`
	CreatedAt time.Time `json:"created_at"`
}

func (x *Link) ZScore() int {
	return int(x.CreatedAt.Unix())
}

func (x *Link) EncodeRedisString() string {
	linkBytes, _ := json.Marshal(x)
	return string(linkBytes)
}

func (x *Link) DecodeRedisString(src string) {
	json.Unmarshal([]byte(src), x)
}

type Links []Link

func (x *Links) Key() string { return (*x)[0].ObjectKey }
func (x *Links) NewKey(key string) {
	*x = append([]Link{{
		ObjectKey: key,
		CreatedAt: time.Now(),
	}}, *x...)
}

func (x Part) String() string {
	return fmt.Sprintf("Project: %s Part: %s", x.Project, strings.Title(x.Name))
}

func (x Part) SheetLink(bucket string) string {
	if bucket == "" || len(x.Sheets) == 0 {
		return "#"
	} else {
		return fmt.Sprintf("/download?bucket=%s&object=%s", bucket, x.Sheets[0].ObjectKey)
	}
}

func (x Part) ClickLink(bucket string) string {
	if bucket == "" || len(x.Clix) == 0 {
		return "#"
	} else {
		return fmt.Sprintf("/download?bucket=%s&object=%s", bucket, x.Clix.Key())
	}
}

func (x Part) Validate() error {
	switch true {
	case projects.Exists(x.Project) == false:
		return projects.ErrNotFound
	case ValidNames(x.Name) == false:
		return ErrInvalidPartName
	default:
		return nil
	}
}

func ValidNames(names ...string) bool {
	for _, name := range names {
		if name == "" {
			return false
		}
	}
	return true
}

type ID struct { // this can be a redis hash
	Project string `json:"project" redis:"project"`
	Name    string `json:"part_name" redis:"part_name"`
}
