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

	redis.InitializeFromEnv()
}

var lrand = rand.New(rand.NewSource(time.Now().UnixNano()))

func newNamespace() string { return "testing" + strconv.Itoa(lrand.Int()) }

func newSessions() *login.Store {
	return login.NewStore(newNamespace(), login.Config{
		CookieName: "vvgo-test-cookie",
		CookiePath: "/",
	})
}
