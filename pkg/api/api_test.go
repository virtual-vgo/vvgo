package api

import (
	"github.com/virtual-vgo/vvgo/pkg/redis"
)

func init() {
	PublicFiles = "../../public"
	redis.InitializeFromEnv()
}
