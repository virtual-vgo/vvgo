package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCache_GetObject(t *testing.T) {
	t.Run("doesnt exist", func(t *testing.T) {
		cache := NewCache(CacheOpts{})
		var gotObject Object
		err := cache.GetObject(context.Background(), "test-file", &gotObject)
		assert.Equal(t, err, ErrObjectNotFound)
	})
	t.Run("exists", func(t *testing.T) {
		cache := NewCache(CacheOpts{})
		cache.cache["test-file"] = Object{ContentType: "test-media"}
		var gotObject Object
		err := cache.GetObject(context.Background(), "test-file", &gotObject)
		assert.NoError(t, err)
		assert.Equal(t, Object{ContentType: "test-media"}, gotObject)
	})
}

func TestCache_PutObject(t *testing.T) {
	ctx := context.Background()
	cache := NewCache(CacheOpts{})
	require.NoError(t, cache.PutObject(ctx, "test-file", &Object{ContentType: "test-media"}))
	assert.Equal(t, Object{ContentType: "test-media"}, cache.cache["test-file"])
}
