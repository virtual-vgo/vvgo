package storage

import (
	"context"
	"errors"
)

// Cache is an in-memory object cache.
// This can serve as a cache proxy between object storage, or as a stand-alone cache.
type Cache struct {
	cache  map[string]Object // the cache map itself
	bucket *Bucket           // optional bucket storage
}

type CacheOpts struct {
	// If this is not nil, each write to the cache will additionally be written to a file in this bucket.
	Bucket *Bucket
}

func NewCache(opts CacheOpts) *Cache {
	return &Cache{
		bucket: opts.Bucket,
		cache:  make(map[string]Object),
	}
}

var ErrObjectNotFound = errors.New("object not found")

// GetObject will first try to find the object in the cache.
// If it's a miss and x.Bucket is not nil, this function will query the bucket and store the results in the cache.
func (x *Cache) GetObject(ctx context.Context, name string, dest *Object) error {
	// try to read from the map
	var ok bool
	*dest, ok = x.cache[name]
	if ok {
		return nil
	}

	// try to read from the bucket
	if x.bucket == nil {
		return ErrObjectNotFound
	}
	if err := x.bucket.GetObject(ctx, name, dest); err != nil {
		return nil
	}
	x.cache[name] = *dest
	return nil
}

// PutObject puts an object into the cache.
// If bucket is not nil, the file will also be written to the bucket.
func (x *Cache) PutObject(ctx context.Context, name string, object *Object) error {
	if x.bucket != nil {
		if err := WithBackup(x.bucket.PutObject)(ctx, name, object); err != nil {
			return err
		}
	}

	x.cache[name] = *object
	return nil
}
