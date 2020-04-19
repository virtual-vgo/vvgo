package storage

import (
	"context"
	"errors"
	"github.com/virtual-vgo/vvgo/pkg/locker"
)

// In-memory caching between bucket operations
type Cache struct {
	bucket *Bucket
	cache  map[string]Object
	locker *locker.Locker
}

func NewCache(bucket *Bucket) *Cache {
	return &Cache{
		bucket: bucket,
		cache:  nil,
		locker: nil,
	}
}

var ErrObjectNotFound = errors.New("object not found")

func (x *Cache) StatObject(ctx context.Context, objectName string, dest *Object) error {
	x.locker.Lock(ctx)
	defer x.locker.Unlock(ctx)

	_, ok := x.cache[objectName]
	if ok {
		*dest = x.cache[objectName]
		return nil
	}

	if x.bucket != nil {
		return x.bucket.StatObject(ctx, objectName, dest)
	}
	return ErrObjectNotFound
}

func (x *Cache) GetObject(ctx context.Context, objectName string, dest *Object) error {
	x.locker.Lock(ctx)
	defer x.locker.Unlock(ctx)

	_, ok := x.cache[objectName]
	if ok {
		*dest = x.cache[objectName]
		return nil
	}

	if x.bucket != nil {
		return x.bucket.GetObject(ctx, objectName, dest)
	}
	return ErrObjectNotFound
}

func (x *Cache) PutObject(ctx context.Context, name string, object *Object) error {
	x.locker.Lock(ctx)
	defer x.locker.Unlock(ctx)

	if x.bucket != nil {
		if err := x.bucket.PutObject(ctx, name, object); err != nil {
			return err
		}
	}

	x.cache[name] = *object
	return nil
}
