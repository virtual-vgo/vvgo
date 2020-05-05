package storage

import (
	"context"
	"github.com/bsm/redislock"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"sync"
	"time"
)

const RedisLockDeadline = 5 * 60 * time.Second

// RedisLocker provides a mutex api over redis.
type RedisLocker struct {
	redisKey  string
	redisLock *redislock.Lock
	client    *redislock.Client
}

func (x *RedisClient) NewLocker(key string) *RedisLocker {
	return &RedisLocker{redisKey: key, client: redislock.New(x.client)}
}

// Lock obtains a lock in redis.
func (x *RedisLocker) Lock(ctx context.Context) error {
	ctx, span := x.newSpan(ctx, "RedisLocker.Lock")
	defer span.Send()

	opts := redislock.Options{
		Context: ctx,
	}
	redisLock, err := x.client.Obtain(x.redisKey, RedisLockDeadline, &opts)
	if err != nil {
		span.AddField("error", err)
		return err
	}
	x.redisLock = redisLock
	return nil
}

// Unlock releases the lock if it exists.
func (x *RedisLocker) Unlock(ctx context.Context) {
	if x.redisLock != nil {
		_, span := x.newSpan(ctx, "RedisLocker.Unlock")
		defer span.Send()
		if err := x.redisLock.Release(); err != nil {
			logger.WithError(err).Error("redislock.Release() failed")
			span.AddField("error", err)
		}
	}
}

func (x *RedisLocker) newSpan(ctx context.Context, name string) (context.Context, tracing.Span) {
	ctx, span := tracing.StartSpan(ctx, name)
	tracing.AddField(ctx, "locker_key", x.redisKey)
	return ctx, span
}

// MemLocker mimics the redis locker api, but only using a sync.Mutex
type MemLocker struct{ lock sync.Mutex }

func (x *MemLocker) Lock(_ context.Context) error { x.lock.Lock(); return nil }
func (x *MemLocker) Unlock(_ context.Context)     { x.lock.Unlock() }
