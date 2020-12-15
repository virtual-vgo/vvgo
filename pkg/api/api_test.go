package api

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"math/rand"
	"strconv"
	"time"
)

func init() {
	PublicFiles = "../../public"

	var redisConfig redis.Config
	envconfig.MustProcess("REDIS", &redisConfig)
	redis.Initialize(redisConfig)
}

var lrand = rand.New(rand.NewSource(time.Now().UnixNano()))

func newNamespace() string { return "testing" + strconv.Itoa(lrand.Int()) }

func newSessions() *login.Store {
	return login.NewStore(newNamespace(), login.Config{
		CookieName: "vvgo-test-cookie",
		CookiePath: "/",
	})
}
