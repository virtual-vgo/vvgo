package redis

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestConfig_ParseEnv(t *testing.T) {
	envs := map[string]string{
		"REDIS_ADDRESS":  "redis-address",
		"REDIS_NETWORK":  "redis-network",
		"REDIS_POOLSIZE": "17",
	}
	want := Config{
		Network:  "redis-network",
		Address:  "redis-address",
		PoolSize: 17,
	}

	for k, v := range envs {
		require.NoError(t, os.Setenv(k, v))
	}
	var got Config
	envconfig.Usage("REDIS", &got)
	require.NoError(t, envconfig.Process("REDIS", &got))
	assert.Equal(t, want, got)
}
