package api

import (
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"math/rand"
	"strconv"
	"time"
)

func init() {
	PublicFiles = "../../public"
}

var lrand = rand.New(rand.NewSource(time.Now().UnixNano()))

func newSessions(cookieDomain string) *login.Store {
	return login.NewStore("testing"+strconv.Itoa(lrand.Int()), login.Config{
		CookieName:   "vvgo-test-cookie",
		CookieDomain: cookieDomain,
		CookiePath:   "/",
	})
}

func newParts() *parts.RedisParts {
	return parts.NewParts("testing" + strconv.Itoa(lrand.Int()))
}
