package minio

import (
	"context"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"testing"
)

func init() {
	var redisConfig redis.Config
	envconfig.MustProcess("REDIS", &redisConfig)
	redis.Initialize(redisConfig)
}

func TestReadConfig(t *testing.T) {
	var got Config
	got.readDefaults()
	err := got.readRedis(context.Background())
	require.NoError(t, err)
	fmt.Printf("%#v", got)
}
