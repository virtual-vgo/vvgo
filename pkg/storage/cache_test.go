package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemCache_Get(t *testing.T) {
	ctx := context.Background()
	warehouse, err := NewWarehouse(Config{NoOp: true})
	require.NoError(t, err)
	noOpBucket, err := warehouse.NewBucket(ctx, "test-bucket")
	require.NoError(t, err)

	t.Run("doesnt exist", func(t *testing.T) {
		cache := &MemCache{}
		var gotObject Object
		err := cache.GetObject(context.Background(), "test-file", &gotObject)
		assert.Equal(t, err, ErrObjectNotFound)
	})
	t.Run("exists/no bucket", func(t *testing.T) {
		cache := NewCache(CacheOpts{Bucket: noOpBucket})
		cache.cache["test-file"] = Object{ContentType: "test-media"}
		var gotObject Object
		err := cache.GetObject(context.Background(), "test-file", &gotObject)
		assert.NoError(t, err)
		assert.Equal(t, Object{ContentType: "test-media"}, gotObject)
	})
	t.Run("exists/bucket", func(t *testing.T) {
		cache := NewCache(CacheOpts{Bucket: noOpBucket})
		var gotObject Object
		err := cache.GetObject(context.Background(), "test-file", &gotObject)
		assert.NoError(t, err)
		assert.Equal(t, Object{}, gotObject)
	})
}

func TestCache_PutObject(t *testing.T) {
	ctx := context.Background()
	warehouse, err := NewWarehouse(Config{NoOp: true})
	require.NoError(t, err)
	noOpBucket, err := warehouse.NewBucket(ctx, "test-bucket")
	require.NoError(t, err)

	t.Run("no bucket", func(t *testing.T) {
		cache := NewCache(CacheOpts{})
		require.NoError(t, cache.PutObject(ctx, "test-file", &Object{ContentType: "test-media"}))
		assert.Equal(t, Object{ContentType: "test-media"}, cache.cache["test-file"])
	})

	t.Run("bucket", func(t *testing.T) {
		cache := NewCache(CacheOpts{noOpBucket})
		require.NoError(t, cache.PutObject(ctx, "test-file", &Object{ContentType: "test-media"}))
		assert.Equal(t, Object{ContentType: "test-media"}, cache.cache["test-file"])
	})
}
