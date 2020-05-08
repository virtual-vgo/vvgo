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
	ErrInvalidPartName   = fmt.Errorf("invalid part name")
	ErrInvalidPartNumber = fmt.Errorf("invalid part number")
)

type RedisParts struct {
	namespace string
	pool      *redis.Client
}

// List returns a slice of all parts in the database.
// If the first request to redis fails, this function will return an error.
// Subsequent errors will only be logged.
func (x *RedisParts) List(ctx context.Context) ([]Part, error) {
	localCtx, span := tracing.StartSpan(ctx, "RedisParts.List()")
	defer span.Send()

	// Read the all the part _keys_ first.
	var partKeys []string
	if err := x.pool.Do(localCtx, redis.Cmd(&partKeys, "ZRANGE", x.namespace+":parts:index", "0", "-1")); err != nil {
		return nil, err
	}

	// Now we can read each individual part.
	// Radix will automatically batch/pipeline requests for us, so we can queue up a bunch of requests in go routines.
	// https://pkg.go.dev/github.com/mediocregopher/radix/v3?tab=doc#hdr-Implicit_pipelining
	parts := make([]Part, len(partKeys))
	var wg sync.WaitGroup
	wg.Add(len(partKeys))
	for i := range partKeys {
		go func(i int) {
			defer wg.Done()
			x.readPart(ctx, partKeys[i], &parts[i])
		}(i)
	}
	wg.Wait()
	return parts, nil
}

func (x *RedisParts) readPart(ctx context.Context, key string, dest *Part) {
	dest.DecodeRedisKey(key)
	sheetsKey := x.namespace + ":parts:" + key + ":sheets"
	dest.Sheets = x.readLinks(ctx, sheetsKey)
	clixKey := x.namespace + ":parts:" + key + ":clix"
	dest.Clix = x.readLinks(ctx, clixKey)
}

func (x *RedisParts) readLinks(ctx context.Context, key string) []Link {
	var raw []string
	if err := x.pool.Do(ctx, redis.Cmd(&raw, "ZREVRANGE", key, "0", "-1")); err != nil {
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

	// Validate the parts.
	for i := range parts {
		if err := parts[i].Validate(); err != nil {
			return err
		}
	}

	args := make([]string, 0, 1+2*len(parts))
	args = append(args, x.namespace+":parts:index")
	for i := range parts {
		score := strconv.Itoa(parts[i].ZScore())
		member := parts[i].RedisKey()
		args = append(args, score, member)
	}
	if err := x.pool.Do(ctx, redis.Cmd(nil, "ZADD", args...)); err != nil {
		logger.WithError(err).WithField("args", args).Error("ZADD")
	}

	// Like with List(), we'll leverage pipelining and queue up lots of goroutines.
	var wg sync.WaitGroup
	wg.Add(len(parts))
	for i := range parts {
		go func(src *Part) {
			defer wg.Done()
			x.savePart(localCtx, src)
		}(&parts[i])
	}
	wg.Wait()
	return nil
}

func (x *RedisParts) savePart(ctx context.Context, src *Part) {
	score := strconv.Itoa(src.ZScore())
	partsKey := x.namespace + ":parts:index"
	if err := x.pool.Do(ctx, redis.Cmd(nil, "ZADD", partsKey, score, src.RedisKey())); err != nil {
		logger.WithError(err).Error("ZADD")
	}
	sheetsKey := x.namespace + ":parts:" + src.RedisKey() + ":sheets"
	x.saveLinks(ctx, sheetsKey, src.Sheets)
	clixKey := x.namespace + ":parts:" + src.RedisKey() + ":clix"
	x.saveLinks(ctx, clixKey, src.Clix)
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
	if err := x.pool.Do(ctx, redis.Cmd(nil, "ZADD", args...)); err != nil {
		logger.WithError(err).WithField("args", args).Error("ZADD")
	}
}

type Part struct {
	ID
	Clix   Links `json:"click,omitempty"`
	Sheets Links `json:"sheet,omitempty"`
}

func (x *Part) RedisKey() string {
	return fmt.Sprintf("%s:%s:%d", x.ID.Project, x.ID.Name, x.ID.Number)
}

func (x *Part) DecodeRedisKey(str string) {
	got := strings.SplitN(str, ":", 3)
	x.ID.Project = got[0]
	x.ID.Name = got[1]
	num, _ := strconv.Atoi(got[2])
	x.ID.Number = uint8(num)
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
	return fmt.Sprintf("Project: %s Part: %s #%d", x.Project, strings.Title(x.Name), x.Number)
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
	case ValidNumbers(x.Number) == false:
		return ErrInvalidPartNumber
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

func ValidNumbers(numbers ...uint8) bool {
	for _, n := range numbers {
		if n == 0 {
			return false
		}
	}
	return true
}

type ID struct { // this can be a redis hash
	Project string `json:"project" redis:"project"`
	Name    string `json:"part_name" redis:"part_name"`
	Number  uint8  `json:"part_number" redis:"part_number"`
}

func (id ID) String() string {
	return fmt.Sprintf("%s-%s-%d", id.Project, id.Name, id.Number)
}
