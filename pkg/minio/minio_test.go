package minio

import (
	"github.com/virtual-vgo/vvgo/pkg/redis"
)

func init() {
	redis.InitializeFromEnv()
}
