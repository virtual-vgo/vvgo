package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/locker"
)

// In-memory caching between bucket operations
type Cache struct {
	bucket *Bucket
	cache  map[string]Object
	locker *locker.Locker
}

type CacheOpts struct {
	Bucket   *Bucket
	RedisKey string
}

func NewCache(opts CacheOpts) *Cache {
	return &Cache{
		bucket: opts.Bucket,
		cache:  make(map[string]Object),
		locker: locker.NewLocker(locker.Opts{RedisKey: opts.RedisKey}),
	}
}

var ErrObjectNotFound = errors.New("object not found")

func (x *Cache) GetObject(ctx context.Context, name string, dest *Object) error {
	if err := x.locker.Lock(ctx); err != nil {
		return fmt.Errorf("locker.Lock() failed: %v", err)
	}
	defer x.locker.Unlock(ctx)

	switch err := x.readObject(name, dest); err {
	case nil:
		return nil
	case ErrObjectNotFound:
		if x.bucket == nil {
			return ErrObjectNotFound
		}

		// stat the object
		if err := x.bucket.GetObject(ctx, name, dest); err != nil {
			return nil
		}
		x.cache[name] = *dest
		return nil
	default:
		return err
	}
}

func (x *Cache) readObject(name string, dest *Object) error {
	// check if the object is already in cache
	if _, ok := x.cache[name]; ok {
		*dest = x.cache[name]
		return nil
	} else {
		return ErrObjectNotFound
	}
}

func (x *Cache) PutObject(ctx context.Context, name string, object *Object) error {
	if err := x.locker.Lock(ctx); err != nil {
		return err
	}
	defer x.locker.Unlock(ctx)

	cacheBuffer := object.Buffer
	if x.bucket != nil {
		if err := WithBackup(x.bucket.PutObject)(ctx, name, object); err != nil {
			return err
		}
	}

	x.cache[name] = *NewObject(object.ContentType, object.Tags, &cacheBuffer)
	return nil
}
