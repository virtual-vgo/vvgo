package locker

import (
	"context"
	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"sync"
	"time"
)

const RedisLockDeadline = 5 * 60 * time.Second

var logger = log.Logger()

// Locksmith is used to build new mutex lockers.
type Locksmith struct {
	config      Config
	redisClient *redis.Client
}

type Config struct {
	// Enabled establishing locks with redis.
	RedisEnabled bool `split_words:"true"`

	// Tcp address of the redis server.
	RedisAddress string `split_words:"true"`
}

func NewLocksmith(config Config) *Locksmith {
	var redisClient *redis.Client
	if config.RedisEnabled {
		redisClient = redis.NewClient(&redis.Options{
			Network: "tcp",
			Addr:    config.RedisAddress,
		})
	}
	return &Locksmith{
		config:      config,
		redisClient: redisClient,
	}
}

// A mutex locker
type Locker struct {
	lock            sync.Mutex
	redisKey        string
	redisLock       *redislock.Lock
	redisLockClient *redislock.Client
}

type Opts struct {
	RedisKey string
}

func (x *Locksmith) NewLocker(opts Opts) *Locker {
	locker := &Locker{redisKey: opts.RedisKey}
	if x.redisClient != nil {
		locker.redisLockClient = redislock.New(x.redisClient)
	}
	return locker
}

func (x *Locker) Lock(ctx context.Context) error {
	x.lock.Lock()
	if x.redisLockClient != nil {
		ctx, span := x.newSpan(ctx, "Locker.Lock")
		defer span.Send()

		opts := redislock.Options{
			Context: ctx,
		}
		redisLock, err := x.redisLockClient.Obtain(x.redisKey, RedisLockDeadline, &opts)
		if err != nil {
			span.AddField("error", err)
			return err
		} else {
			x.redisLock = redisLock
			return nil
		}
	}
	return nil
}

func (x *Locker) Unlock(ctx context.Context) {
	defer x.lock.Unlock()
	if x.redisLock != nil {
		_, span := x.newSpan(ctx, "Locker.Unlock")
		defer span.Send()
		if err := x.redisLock.Release(); err != nil {
			logger.WithError(err).Error("redislock.Release() failed")
			span.AddField("error", err)
		}
	}
}

func (x *Locker) newSpan(ctx context.Context, name string) (context.Context, tracing.Span) {
	ctx, span := tracing.StartSpan(ctx, name)
	tracing.AddField(ctx, "locker_key", x.redisKey)
	return ctx, span
}
