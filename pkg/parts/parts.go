package parts

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mediocregopher/radix/v3"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/projects"
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
	pool      *radix.Pool
}

func logOnError(err error, msg string) {
	if err != nil {
		logger.WithError(err).Error(msg)
	}
}

func (x *RedisParts) List(ctx context.Context) ([]Part, error) {
	_, span := tracing.StartSpan(ctx, "RedisParts.List()")
	defer span.Send()

	var partKeys []string
	if err := x.pool.Do(radix.Cmd(&partKeys, "ZRANGE", x.namespace+":parts:index", "0", "-1")); err != nil {
		return nil, err
	}

	parts := make([]Part, len(partKeys))

	var wg sync.WaitGroup
	wg.Add(len(partKeys))
	for i := range partKeys {
		go func(i int) {
			defer wg.Done()
			var part Part
			part.DecodeRedisKey(partKeys[i])
			sheetsKey := x.namespace + ":parts:" + partKeys[i] + ":sheets"
			clixKey := x.namespace + ":parts:" + partKeys[i] + ":clix"
			var raw []string
			logOnError(x.pool.Do(radix.Cmd(&raw, "ZREVRANGE", sheetsKey, "0", "-1")), "ZRANGE")
			part.Sheets = make([]Link, len(raw))
			for i := range raw {
				part.Sheets[i].DecodeRedisString(raw[i])
			}
			raw = nil
			logOnError(x.pool.Do(radix.Cmd(&raw, "ZREVRANGE", clixKey, "0", "-1")), "ZRANGE")
			part.Clix = make([]Link, len(raw))
			for i := range raw {
				part.Clix[i].DecodeRedisString(raw[i])
			}
			parts[i] = part
		}(i)
	}
	wg.Wait()
	return parts, nil
}

func (x *RedisParts) Save(ctx context.Context, parts []Part) error {
	_, span := tracing.StartSpan(ctx, "RedisParts.Save")
	defer span.Send()

	// first, validate all the parts
	for i := range parts {
		if err := parts[i].Validate(); err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(parts))
	for i := range parts {
		go func(i int) {
			defer wg.Done()
			score := strconv.Itoa(parts[i].ZScore())
			partsKey := x.namespace + ":parts:index"
			sheetsKey := x.namespace + ":parts:" + parts[i].RedisKey() + ":sheets"
			clixKey := x.namespace + ":parts:" + parts[i].RedisKey() + ":clix"

			logOnError(x.pool.Do(radix.Cmd(nil, "ZADD", partsKey, score, parts[i].RedisKey())), "ZADD")

			for j := range parts[i].Sheets {
				score := strconv.Itoa(parts[i].Sheets[j].ZScore())
				member := parts[i].Sheets[j].EncodeRedisString()
				logOnError(x.pool.Do(radix.Cmd(nil, "ZADD", sheetsKey, score, member)), "ZADD")
			}

			for j := range parts[i].Clix {
				score := strconv.Itoa(parts[i].Clix[j].ZScore())
				member := parts[i].Clix[j].EncodeRedisString()
				logOnError(x.pool.Do(radix.Cmd(nil, "ZADD", clixKey, score, member)), "ZADD")
			}
		}(i)
	}
	wg.Wait()
	return nil
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
