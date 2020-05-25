package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"os"
	"testing"
)

func TestConfig_ParseEnv(t *testing.T) {
	envs := map[string]string{
		"TRACING_HONEYCOMB_DATASET":   "tracing-honeycomb-dataset",
		"TRACING_HONEYCOMB_WRITE_KEY": "tracing-honeycomb-write-key",
		"TRACING_SERVICE_NAME":        "tracing-service-name",
		"API_LISTEN_ADDRESS":          "listen-address",
		"API_DISTRO_BUCKET_NAME":      "distro-bucket-name",
		"API_BACKUPS_BUCKET_NAME":     "backups-bucket-name",
		"API_REDIS_NAMESPACE":         "redis-namespace",
		"API_MEMBER_USER":             "member-user",
		"API_MEMBER_PASS":             "member-pass",
		"REDIS_ADDRESS":               "redis-address",
		"REDIS_NETWORK":               "redis-network",
		"REDIS_POOL_SIZE":             "17",
		"MINIO_ENDPOINT":              "minio-endpoint",
		"MINIO_REGION":                "minio-region",
		"MINIO_ACCESSKEY":             "minio-access-key",
		"MINIO_SECRETKEY":             "minio-secret-key",
		"MINIO_USESSL":                "true",
		"API_UPLOADER_TOKEN":          "uploader-token",
		"API_DEVELOPER_TOKEN":         "developer-token",
		"API_LOGIN_COOKIE_NAME":       "login-cookie-name",
		"API_LOGIN_COOKIE_DOMAIN":     "login-cookie-domain",
		"API_LOGIN_COOKIE_PATH":       "login-cookie-path",
	}
	want := Config{
		ApiConfig: api.ServerConfig{
			ListenAddress:     "listen-address",
			MemberUser:        "member-user",
			MemberPass:        "member-pass",
			DistroBucketName:  "distro-bucket-name",
			BackupsBucketName: "backups-bucket-name",
			RedisNamespace:    "redis-namespace",
			DeveloperToken:    "developer-token",
			UploaderToken:     "uploader-token",
			Login: login.Config{
				CookieName:   "login-cookie-name",
				CookieDomain: "login-cookie-domain",
				CookiePath:   "login-cookie-path",
			},
		},
		TracingConfig: tracing.Config{
			HoneycombWriteKey: "tracing-honeycomb-write-key",
			HoneycombDataset:  "tracing-honeycomb-dataset",
			ServiceName:       "tracing-service-name",
		},
		RedisConfig: redis.Config{
			Network:  "redis-network",
			Address:  "redis-address",
			PoolSize: 17,
		},
		MinioConfig: storage.Config{
			Endpoint:  "minio-endpoint",
			Region:    "minio-region",
			AccessKey: "minio-access-key",
			SecretKey: "minio-secret-key",
			UseSSL:    true,
		},
	}

	for k, v := range envs {
		os.Setenv(k, v)
	}
	var got Config
	envconfig.Usage("", &got)
	got.ParseEnv()
	assert.Equal(t, want, got)
}
