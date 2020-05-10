package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewWarehouse(t *testing.T) {
	t.Run("noop=false", func(t *testing.T) {
		t.Run("endpoint=localhost:9000", func(t *testing.T) {
			gotWarehouse, err := NewWarehouse(Config{
				NoOp: false,
				Minio: MinioConfig{
					Endpoint: "localhost:9000",
				},
			})
			require.NoError(t, err)
			require.NotNil(t, gotWarehouse, "warehouse")
			assert.Equal(t, Config{
				NoOp: false,
				Minio: MinioConfig{
					Endpoint: "localhost:9000",
				},
			}, gotWarehouse.config, "warehouse.config")
			assert.NotNil(t, gotWarehouse.minioClient, "warehouse.minioClient")
		})

		t.Run("endpoint=::invalid::endpoint::", func(t *testing.T) {
			gotWarehouse, err := NewWarehouse(Config{
				NoOp: false,
				Minio: MinioConfig{
					Endpoint: "::invalid::endpoint::",
				},
			})
			require.Error(t, err)
			assert.Nil(t, gotWarehouse)
		})
	})

	t.Run("noop=true", func(t *testing.T) {
		for _, endpoint := range []string{"", "localhost:9000", "::invalid::endpoint::"} {
			t.Run("endpoint="+endpoint, func(t *testing.T) {
				gotWarehouse, err := NewWarehouse(Config{
					NoOp: true,
					Minio: MinioConfig{
						Endpoint: endpoint,
					},
				})
				require.NoError(t, err)
				require.NotNil(t, gotWarehouse, "warehouse")
				assert.Equal(t, Config{
					NoOp: true,
					Minio: MinioConfig{
						Endpoint: endpoint,
					},
				}, gotWarehouse.config, "warehouse.config")
				assert.Nil(t, gotWarehouse.minioClient, "warehouse.minioClient")
			})
		}
	})
}

func TestWarehouse_NewBucket(t *testing.T) {
	ctx := context.Background()
	t.Run("noop=false/endpoint=", func(t *testing.T) {
		warehouse, err := NewWarehouse(Config{
			NoOp: false,
			Minio: MinioConfig{
				Endpoint: "localhost:9000",
			},
		})
		require.NoError(t, err)
		gotBucket, err := warehouse.NewBucket(ctx, "test-bucket")
		assert.Error(t, err)
		assert.Nil(t, gotBucket)
	})

	t.Run("noop=true", func(t *testing.T) {
		warehouse, err := NewWarehouse(Config{NoOp: true})
		require.NoError(t, err)
		gotBucket, err := warehouse.NewBucket(ctx, "test-bucket")
		require.NoError(t, err)
		assert.Equal(t, &Bucket{Name: "test-bucket", noOp: true}, gotBucket)
	})
}

func TestFile_ValidateMediaType(t *testing.T) {
	for _, tt := range []struct {
		name      string
		file      File
		pre       string
		wantError error
	}{
		{
			name: "success",
			pre:  "text/html",
			file: File{
				MediaType: "text/html",
				Ext:       ".html",
				Bytes:     []byte(`<!doctype html></html>`),
			},
			wantError: nil,
		},
		{
			name: ErrInvalidMediaType.Error(),
			pre:  "text/html",
			file: File{
				MediaType: "text/plain",
				Ext:       ".html",
				Bytes:     []byte(`<!doctype html></html>`),
			},
			wantError: ErrInvalidMediaType,
		},
		{
			name: ErrInvalidFileExtension.Error(),
			pre:  "text/html",
			file: File{
				MediaType: "text/html",
				Ext:       ".txt",
				Bytes:     []byte(`<!doctype html></html>`),
			},
			wantError: ErrInvalidFileExtension,
		},
		{
			name: ErrDetectedInvalidContent.Error(),
			pre:  "text/html",
			file: File{
				MediaType: "text/html",
				Ext:       ".html",
				Bytes:     []byte(``),
			},
			wantError: ErrDetectedInvalidContent,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantError, tt.file.ValidateMediaType(tt.pre))
		})
	}
}

func TestFile_ObjectKey(t *testing.T) {
	wantKey := "c54ee0315feb10edabf66fa45fe1b916.html"
	gotKey := File{MediaType: "text/html", Ext: ".html", Bytes: []byte(`<!doctype html></html>`)}.ObjectKey()
	assert.Equal(t, wantKey, gotKey)
}

func TestNewObject(t *testing.T) {
	wantObject := &Object{
		ContentType: "text/html",
		Tags:        map[string]string{"howdy": "there"},
		Bytes:       []byte(`<!doctype html></html>`),
	}
	gotObject := NewObject(
		"text/html",
		map[string]string{"howdy": "there"},
		[]byte(`<!doctype html></html>`),
	)
	assert.Equal(t, wantObject, gotObject)
}

func TestBucket_NoOp(t *testing.T) {
	ctx := context.Background()
	warehouse, err := NewWarehouse(Config{NoOp: true})
	require.NoError(t, err)
	bucket, err := warehouse.NewBucket(ctx, "test-bucket")
	require.NoError(t, err)
	var object Object
	var file File
	assert.NoError(t, bucket.StatFile(ctx, "test-file", &file), "bucket.StatFile")
	assert.NoError(t, bucket.PutFile(ctx, &file), "bucket.StatFile")
	assert.NoError(t, bucket.StatObject(ctx, "test-object", &object), "bucket.StatObject()")
	assert.NoError(t, bucket.GetObject(ctx, "test-object", &object), "bucket.StatObject()")
	assert.NoError(t, bucket.PutObject(ctx, "test-object", &object), "bucket.StatObject()")
	gotUrl, gotErr := bucket.DownloadURL(ctx, "test-object")
	assert.NoError(t, gotErr, "bucket.DownloadURL")
	assert.Equal(t, "#", gotUrl)
}
