package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"os"
	"testing"
)

func TestConfig_ParseEnv(t *testing.T) {
	envs := map[string]string{
		"INITIALIZE_STORAGE": "true",
		"MINIO_ENDPOINT":     "minio-endpoint",
		"MINIO_REGION":       "minio-region",
		"MINIO_ACCESS_KEY":   "minio-access-key",
		"MINIO_SECRET_KEY":   "minio-secret-key",
		"MINIO_USE_SSL":      "true",
		"REDIS_ADDRESS":      "redis-address",
		"LISTEN_ADDRESS":     "listen-address",
		"BASIC_AUTH_USER":    "basic-auth-user",
		"BASIC_AUTH_PASS":    "basic-auth-pass",
		"SHEETS_BUCKET_NAME": "sheets-bucket-name",
		"CLIX_BUCKET_NAME":   "clix-bucket-name",
		"PARTS_BUCKET_NAME":  "parts-bucket-name",
		"PARTS_LOCKER_NAME":  "parts-locker-name",
	}
	want := Config{
		InitializeStorage: true,
		StorageConfig: storage.Config{
			MinioConfig: storage.MinioConfig{
				Endpoint:  "minio-endpoint",
				Region:    "minio-region",
				AccessKey: "minio-access-key",
				SecretKey: "minio-secret-key",
				UseSSL:    true,
			},
			RedisConfig: storage.RedisConfig{
				Address: "redis-address",
			},
		},
		ApiConfig: api.ServerConfig{
			ListenAddress:    "listen-address",
			MaxContentLength: 0,
			BasicAuthUser:    "basic-auth-user",
			BasicAuthPass:    "basic-auth-pass",
			SheetsBucketName: "sheets-bucket-name",
			ClixBucketName:   "clix-bucket-name",
			PartsBucketName:  "parts-bucket-name",
			PartsLockerName:  "parts-locker-name",
		},
	}
	for k, v := range envs {
		os.Setenv(k, v)
	}
	var got Config
	got.ParseEnv()
	assert.Equal(t, want, got)
}
