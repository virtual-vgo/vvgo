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
		"INITIALIZE_STORAGE":         "true",
		"STORAGE_MINIO_ENDPOINT":     "minio-endpoint",
		"STORAGE_MINIO_REGION":       "minio-region",
		"STORAGE_MINIO_ACCESSKEY":    "minio-access-key",
		"STORAGE_MINIO_SECRETKEY":    "minio-secret-key",
		"STORAGE_MINIO_USESSL":       "true",
		"STORAGE_REDIS_ADDRESS":      "redis-address",
		"API_LISTEN_ADDRESS":         "listen-address",
		"API_MAX_CONTENT_LENGTH":     "1000000",
		"API_SHEETS_BUCKET_NAME":     "sheets-bucket-name",
		"API_CLIX_BUCKET_NAME":       "clix-bucket-name",
		"API_PARTS_BUCKET_NAME":      "parts-bucket-name",
		"API_PARTS_LOCKER_KEY":       "parts-locker-key",
		"API_MEMBER_BASIC_AUTH_USER": "member-basic-auth-user",
		"API_MEMBER_BASIC_AUTH_PASS": "member-basic-auth-pass",
		"API_PREP_REP_TOKEN":         "prep-rep-token",
		"API_ADMIN_TOKEN":            "admin-token",
	}
	want := Config{
		InitializeStorage: true,
		StorageConfig: storage.Config{
			Minio: storage.MinioConfig{
				Endpoint:  "minio-endpoint",
				Region:    "minio-region",
				AccessKey: "minio-access-key",
				SecretKey: "minio-secret-key",
				UseSSL:    true,
			},
			Redis: storage.RedisConfig{
				Address: "redis-address",
			},
		},
		ApiConfig: api.ServerConfig{
			ListenAddress:       "listen-address",
			MaxContentLength:    1e6,
			SheetsBucketName:    "sheets-bucket-name",
			ClixBucketName:      "clix-bucket-name",
			PartsBucketName:     "parts-bucket-name",
			PartsLockerKey:      "parts-locker-key",
			AdminToken:          "admin-token",
			PrepRepToken:        "prep-rep-token",
			MemberBasicAuthUser: "member-basic-auth-user",
			MemberBasicAuthPass: "member-basic-auth-pass",
		},
	}
	for k, v := range envs {
		os.Setenv(k, v)
	}
	var got Config
	got.ParseEnv()
	assert.Equal(t, want, got)
}
