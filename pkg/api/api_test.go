package api

import (
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"math/rand"
	"strconv"
	"time"
)

func init() {
	PublicFiles = "../../public"
}

var lrand = rand.New(rand.NewSource(time.Now().UnixNano()))

func newParts() *parts.RedisParts {
	return parts.NewParts("testing" + strconv.Itoa(lrand.Int()))
}
