package api

import (
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"math/rand"
	"strconv"
	"time"
)

func init() {
	PublicFiles = "../../public"
}

var lrand = rand.New(rand.NewSource(time.Now().UnixNano()))
var redisClient *redis.Client

func newParts() *parts.RedisParts {
	if redisClient == nil {
		var err error
		redisClient, err = redis.NewClient(redis.Config{
			Network:  "tcp",
			Address:  "localhost:6379",
			PoolSize: 10,
		})
		if err != nil {
			logger.Fatal(err)
		}
	}
	return parts.NewParts(redisClient, "testing"+strconv.Itoa(lrand.Int()))
}
