package api

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/virtual-vgo/vvgo/pkg/minio"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/redis"
)

func init() {
	PublicFiles = "../../public"
	redis.InitializeFromEnv()
	parse_config.UseTestNamespace()
	var minioConfig minio.Config
	envconfig.MustProcess("MINIO", &minioConfig)
	parse_config.WriteToRedisHash(context.Background(), "minio", &minioConfig)
}
