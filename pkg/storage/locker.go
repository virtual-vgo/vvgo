package storage

import (
	"context"
	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
	"time"
)

const Deadline = 5 * 60 * time.Second

type RedisLocker struct {
	RedisConfig
	*redislock.Client
}

type RedisConfig struct {
	Address string
}

func NewRedisLocker(config RedisConfig) *RedisLocker {
	client := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    config.Address,
	})

	return &RedisLocker{
		RedisConfig: config,
		Client:      redislock.New(client),
	}
}

func (x *RedisLocker) Lock(ctx context.Context, key string) *redislock.Lock {
	lock, err := x.Obtain(key, Deadline, &redislock.Options{
		Context: ctx,
	})
	if err != nil {
		logger.WithError(err).Error("redislock.Obtain() failed")
		return nil
	} else {
		return lock
	}
}
