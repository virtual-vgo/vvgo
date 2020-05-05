package storage

import (
	"context"
	"encoding"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"sync"
)

type RedisHash struct {
	Name string
	*redis.Client
}

var ErrKeyIsEmpty = errors.New("key is empty")

func (x *RedisHash) Keys(ctx context.Context) ([]string, error) {
	ctx, span := tracing.StartSpan(ctx, "RedisHash.Keys")
	defer span.Send()
	return x.Client.HKeys(x.Name).Result()
}

func (x *RedisHash) Get(ctx context.Context, name string, dest encoding.BinaryUnmarshaler) error {
	ctx, span := tracing.StartSpan(ctx, "RedisHash.Get")
	defer span.Send()
	destBytes, err := x.Client.WithContext(ctx).HGet(x.Name, name).Bytes()
	switch true {
	case err != nil:
		return err
	case len(destBytes) == 0:
		return ErrKeyIsEmpty
	default:
		return dest.UnmarshalBinary(destBytes)
	}
}

func (x *RedisHash) Set(ctx context.Context, name string, src encoding.BinaryMarshaler) error {
	ctx, span := tracing.StartSpan(ctx, "RedisHash.Set")
	defer span.Send()
	return x.Client.WithContext(ctx).HSet(x.Name, name, src, 0).Err()
}

type MemCache struct {
	Map  map[string][]byte
	lock sync.RWMutex
}

func (x *MemCache) Keys(_ context.Context) ([]string, error) {
	x.lock.RLock()
	defer x.lock.RUnlock()

	keys := make([]string, 0, len(x.Map))
	for key, _ := range x.Map {
		keys = append(keys, key)
	}
	return keys, nil
}

func (x *MemCache) Get(_ context.Context, name string, dest encoding.BinaryUnmarshaler) error {
	x.lock.RLock()
	defer x.lock.RUnlock()
	if x.Map == nil {
		return nil
	}
	switch {
	case x.Map == nil:
		return nil
	case len(x.Map[name]) == 0:
		return ErrKeyIsEmpty
	default:
		return dest.UnmarshalBinary(x.Map[name])
	}

}

func (x *MemCache) Set(_ context.Context, name string, src encoding.BinaryMarshaler) error {
	x.lock.Lock()
	defer x.lock.Unlock()
	if x.Map == nil {
		x.Map = make(map[string][]byte)
	}
	got, err := src.MarshalBinary()
	if err != nil {
		return fmt.Errorf("src.MarshalBinary() failed: %w", err)
	}
	x.Map[name] = got
	return nil
}
