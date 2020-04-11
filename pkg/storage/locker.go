package storage

import (
	"context"
	"github.com/bsm/redislock"
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

func (x *Locker) Lock(ctx context.Context) bool {
	x.lock.Lock()
	lock, err := x.client.Obtain(x.key, RedisLockDeadline, &redislock.Options{
		Context: ctx,
	})
	if err != nil {
		logger.WithError(err).Error("redislock.Obtain() failed")
		return false
	} else {
		x.redisLock = lock
		return true
	}
}

func (x *Locker) Unlock() {
	if err := x.redisLock.Release(); err != nil {
		logger.WithError(err).Error("redislock.Release() failed")
	}
	x.lock.Unlock()
}
