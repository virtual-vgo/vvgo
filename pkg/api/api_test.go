package api

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func init() {
	PublicFiles = "../../public"

	var redisConfig redis.Config
	envconfig.MustProcess("REDIS", &redisConfig)
	redis.Initialize(redisConfig)

	var minioConfig storage.Config
	envconfig.MustProcess("MINIO", &minioConfig)
	storage.Initialize(minioConfig)
}

var lrand = rand.New(rand.NewSource(time.Now().UnixNano()))

func newNamespace() string { return "testing" + strconv.Itoa(lrand.Int()) }

func newSessions() *login.Store {
	return login.NewStore(newNamespace(), login.Config{
		CookieName: "vvgo-test-cookie",
		CookiePath: "/",
	})
}

func newBucket(t *testing.T) *storage.Bucket {
	bucket, err := storage.NewBucket(context.Background(), newNamespace())
	require.NoError(t, err, "storage.NewBucket()")
	return bucket
}
