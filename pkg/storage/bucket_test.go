package storage

import (
	"bytes"
	"context"
	"github.com/kelseyhightower/envconfig"
	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func init() {
	var config Config
	envconfig.MustProcess("MINIO", &config)
	Initialize(config)

}

var localRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func newBucket(t *testing.T) *Bucket {
	bucket, err := NewBucket(context.Background(), "testing"+strconv.Itoa(localRand.Int()))
	require.NoError(t, err, "storage.NewBucket()")
	return bucket
}

func TestNewBucket(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		bucketName := "testing" + strconv.Itoa(localRand.Int())
		gotBucket, err := NewBucket(ctx, bucketName)
		require.NoError(t, err, "NewBucket()")
		assert.Equal(t, &Bucket{
			Name:        bucketName,
			minioRegion: warehouse.config.Region,
			minioClient: warehouse.minioClient,
		}, gotBucket)
	})
	t.Run("failure", func(t *testing.T) {
		ctx := context.Background()
		_, err := NewBucket(ctx, "")
		require.Error(t, err, "NewBucket()")
	})

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

func TestBucket_StatFile(t *testing.T) {
	ctx := context.Background()
	bucket := newBucket(t)
	key := "key.json"
	_, err := bucket.minioClient.PutObject(bucket.Name, key, strings.NewReader(`{}`), -1, minio.PutObjectOptions{
		ContentType: "text/plain",
	})
	require.NoError(t, err, "bucket.PutObject() failed")
	var got File
	assert.NoError(t, bucket.StatFile(ctx, key, &got))
	assert.Equal(t, File{
		MediaType: "text/plain",
		Ext:       ".json",
		objectKey: "key.json",
	}, got)
}

func TestBucket_PutFile(t *testing.T) {
	ctx := context.Background()
	bucket := newBucket(t)
	assert.NoError(t, bucket.PutFile(ctx, &File{
		MediaType: "text/plain",
		Ext:       ".json",
		Bytes:     []byte(`{}`),
	}), "bucket.PutFile() failed")

	var got Object
	assert.NoError(t, bucket.GetObject(ctx, "99914b932bd37a50b983c5e7c90ae93b.json", &got))
	assert.Equal(t, Object{
		ContentType: "text/plain",
		Tags:        map[string]string{},
		Bytes:       []byte(`{}`),
	}, got)
}

func TestBucket_StatObject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		bucket := newBucket(t)
		key := "test-key"
		_, err := bucket.minioClient.PutObject(bucket.Name, key, strings.NewReader(`the earth is flat`), -1, minio.PutObjectOptions{
			ContentType:  "text/plain",
			UserMetadata: map[string]string{"facts": "only"},
		})
		require.NoError(t, err, "bucket.PutObject() failed")
		var got Object
		assert.NoError(t, bucket.StatObject(ctx, key, &got))
		assert.Equal(t, Object{
			ContentType: "text/plain",
			Tags:        map[string]string{"Facts": "only"},
		}, got)
	})

	t.Run("failure", func(t *testing.T) {
		ctx := context.Background()
		bucket := newBucket(t)
		key := "test-key"
		require.NoError(t, bucket.PutObject(ctx, key, &Object{
			ContentType: "text/plain",
			Tags:        map[string]string{"facts": "only"},
			Bytes:       []byte(`the earth is flat`),
		}), "bucket.PutObject() failed")
		var got Object
		assert.Error(t, bucket.StatObject(ctx, "", &got))
	})
}

func TestBucket_GetObject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		bucket := newBucket(t)
		key := "test-key"
		_, err := bucket.minioClient.PutObject(bucket.Name, key, strings.NewReader(`the earth is flat`), -1, minio.PutObjectOptions{
			ContentType:  "text/plain",
			UserMetadata: map[string]string{"facts": "only"},
		})
		require.NoError(t, err, "bucket.PutObject() failed")
		var got Object
		assert.NoError(t, bucket.GetObject(ctx, key, &got))
		assert.Equal(t, Object{
			ContentType: "text/plain",
			Tags:        map[string]string{"Facts": "only"},
			Bytes:       []byte(`the earth is flat`),
		}, got)
	})

	t.Run("failure", func(t *testing.T) {
		ctx := context.Background()
		bucket := newBucket(t)
		key := "test-key"
		require.NoError(t, bucket.PutObject(ctx, key, &Object{
			ContentType: "text/plain",
			Tags:        map[string]string{"facts": "only"},
			Bytes:       []byte(`the earth is flat`),
		}), "bucket.PutObject() failed")
		var got Object
		assert.Error(t, bucket.GetObject(ctx, "", &got))
	})
}

func TestBucket_PutObject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		bucket := newBucket(t)
		key := "test-key"
		require.NoError(t, bucket.PutObject(ctx, key, &Object{
			ContentType: "text/plain",
			Tags:        map[string]string{"facts": "only"},
			Bytes:       []byte(`the earth is flat`),
		}), "bucket.PutObject() failed")
		var got Object
		assert.NoError(t, bucket.GetObject(ctx, key, &got))
		assert.Equal(t, Object{
			ContentType: "text/plain",
			Tags:        map[string]string{"Facts": "only"},
			Bytes:       []byte(`the earth is flat`),
		}, got)
	})

	t.Run("failure", func(t *testing.T) {
		ctx := context.Background()
		bucket := newBucket(t)
		key := "test-key"
		require.Error(t, bucket.PutObject(ctx, "", &Object{
			ContentType: "text/plain",
			Tags:        map[string]string{"facts": "only"},
			Bytes:       []byte(`the earth is flat`),
		}), "bucket.PutObject() failed")
		var got Object
		assert.Error(t, bucket.GetObject(ctx, key, &got))
	})
}

func TestBucket_DownloadURL(t *testing.T) {
	ctx := context.Background()
	bucket := newBucket(t)
	key := "test-key"
	require.NoError(t, bucket.PutObject(ctx, key, &Object{
		ContentType: "text/plain",
		Bytes:       []byte(`the earth is flat`),
	}), "bucket.PutObject() failed")

	downloadUrl, err := bucket.DownloadURL(ctx, key)
	require.NoError(t, err, "bucket.DownloadURL")

	resp, err := http.Get(downloadUrl)
	require.NoError(t, err, "http.Get() failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "resp.StatusCode")
	var gotBody bytes.Buffer
	gotBody.ReadFrom(resp.Body)
	assert.Equal(t, `the earth is flat`, gotBody.String())
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
