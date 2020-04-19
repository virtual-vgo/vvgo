package locker

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConfig_ParseEnv(t *testing.T) {
	envs := map[string]string{
		"LOCKER_REDIS_ADDRESS": "redis-address",
	}
	want := Config{
		RedisAddress: "redis-address",
	}
	for k, v := range envs {
		os.Setenv(k, v)
	}
	var got Config
	got.ParseEnv()
	assert.Equal(t, want, got)
}
