package storage

import (
	"context"
	"github.com/bsm/redislock"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"
	"sync"
)

type Locker struct {
	key       string
	lock      sync.Mutex
	redisLock *redislock.Lock
	client    *redislock.Client
}

func (x *Client) NewLocker(key string) *Locker {
	if x.lockers[key] == nil {
		x.lock.Lock()
		defer x.lock.Unlock()
		x.lockers[key] = &Locker{
			key:    key,
			client: redislock.New(x.redisClient),
		}
	}
	return x.lockers[key]
}

func (x *Locker) newSpan(ctx context.Context, name string) (context.Context, *trace.Span) {
	ctx, span := beeline.StartSpan(ctx, name)
	beeline.AddField(ctx, "locker_key", x.key)
	return ctx, span
}

func (x *Locker) Lock(ctx context.Context) bool {
	ctx, span := x.newSpan(ctx, "Locker.Lock")
	defer span.Send()
	x.lock.Lock()
	lock, err := x.client.Obtain(x.key, RedisLockDeadline, &redislock.Options{
		Context: ctx,
	})
	if err != nil {
		logger.WithError(err).Error("redislock.Obtain() failed")
		span.AddField("error", err)
		return false
	} else {
		x.redisLock = lock
		return true
	}
}

func (x *Locker) Unlock(ctx context.Context) {
	ctx, span := x.newSpan(ctx, "Locker.Unlock")
	defer span.Send()
	if err := x.redisLock.Release(); err != nil {
		logger.WithError(err).Error("redislock.Release() failed")
		span.AddField("error", err)
	}
	x.lock.Unlock()
}
