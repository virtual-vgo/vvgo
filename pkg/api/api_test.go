package api

import (
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/redis"
)

func init() {
	PublicFiles = "../../public"
	redis.InitializeFromEnv()
	parse_config.UseTestNamespace()
}
