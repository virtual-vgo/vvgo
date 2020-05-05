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

type RedisClient struct {
	client redis.Client
}

type RedisConfig struct {
	Address string
}

func NewRedisClient(config RedisConfig) *RedisClient {
	return &RedisClient{
		client: *redis.NewClient(&redis.Options{
			Addr: config.Address,
		}),
	}
}

func (x *RedisClient) NewHash(name string) *RedisHash {
	return &RedisHash{
		Name:   name,
		Client: x.client,
	}
}

type RedisHash struct {
	Name string
	redis.Client
}

var ErrKeyIsEmpty = errors.New("key is empty")

func (x *RedisHash) HKeys(ctx context.Context) ([]string, error) {
	ctx, span := tracing.StartSpan(ctx, "RedisHash.HKeys")
	defer span.Send()
	return x.Client.HKeys(x.Name).Result()
}

func (x *RedisHash) HGet(ctx context.Context, name string, dest encoding.BinaryUnmarshaler) error {
	ctx, span := tracing.StartSpan(ctx, "RedisHash.HGet")
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

func (x *RedisHash) HSet(ctx context.Context, name string, src encoding.BinaryMarshaler) error {
	ctx, span := tracing.StartSpan(ctx, "RedisHash.HSet")
	defer span.Send()
	return x.Client.WithContext(ctx).HSet(x.Name, name, src, 0).Err()
}

type MemHash struct {
	Map  map[string][]byte
	lock sync.RWMutex
}

func (x *MemHash) HKeys(_ context.Context) ([]string, error) {
	x.lock.RLock()
	defer x.lock.RUnlock()

	keys := make([]string, 0, len(x.Map))
	for key, _ := range x.Map {
		keys = append(keys, key)
	}
	return keys, nil
}

func (x *MemHash) HGet(_ context.Context, name string, dest encoding.BinaryUnmarshaler) error {
	x.lock.RLock()
	defer x.lock.RUnlock()
	if x.Map == nil {
		return nil
	}
	switch {
	case x.Map == nil:
		return ErrKeyIsEmpty
	case len(x.Map[name]) == 0:
		return ErrKeyIsEmpty
	default:
		return dest.UnmarshalBinary(x.Map[name])
	}

}

func (x *MemHash) HSet(_ context.Context, name string, src encoding.BinaryMarshaler) error {
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

type MarshalString string
type UnmarshalString string

func (x MarshalString) MarshalBinary() ([]byte, error)    { return []byte(x), nil }
func (x *UnmarshalString) UnmarshalBinary(b []byte) error { *x = UnmarshalString(b); return nil }
