package storage

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConfig_ParseEnv(t *testing.T) {
	envs := map[string]string{
		"STORAGE_MINIO_ENDPOINT":      "minio-endpoint",
		"STORAGE_MINIO_REGION":        "minio-region",
		"STORAGE_MINIO_ACCESSKEY":     "minio-access-key",
		"STORAGE_MINIO_SECRETKEY":     "minio-secret-key",
		"STORAGE_MINIO_USESSL":        "true",
	}
	want := Config{
		Minio: MinioConfig{
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
	got.ParseEnv()
	assert.Equal(t, want, got)
}

