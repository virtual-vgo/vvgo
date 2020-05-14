package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"os"
	"testing"
)

func TestConfig_ParseEnv(t *testing.T) {
	envs := map[string]string{
		"VVGO_SECRET":                 "vvgo-secret",
		"INITIALIZE_STORAGE":          "true",
		"TRACING_HONEYCOMB_DATASET":   "tracing-honeycomb-dataset",
		"TRACING_HONEYCOMB_WRITE_KEY": "tracing-honeycomb-write-key",
		"TRACING_SERVICE_NAME":        "tracing-service-name",
		"API_LISTEN_ADDRESS":          "listen-address",
		"API_MAX_CONTENT_LENGTH":      "1000000",
		"API_SHEETS_BUCKET_NAME":      "sheets-bucket-name",
		"API_CLIX_BUCKET_NAME":        "clix-bucket-name",
		"API_TRACKS_BUCKET_NAME":      "tracks-bucket-name",
		"API_BACKUPS_BUCKET_NAME":     "backups-bucket-name",
		"API_REDIS_NAMESPACE":         "redis-namespace",
		"API_MEMBER_USER":             "member-user",
		"API_MEMBER_PASS":             "member-pass",
		"API_PREP_REP_TOKEN":          "prep-rep-token",
		"API_ADMIN_TOKEN":             "admin-token",
	}
	want := Config{
		Secret: "vvgo-secret",
		ApiConfig: api.ServerConfig{
			ListenAddress:     "listen-address",
			MaxContentLength:  1e6,
			AdminToken:        "admin-token",
			PrepRepToken:      "prep-rep-token",
			MemberUser:        "member-user",
			MemberPass:        "member-pass",
			SheetsBucketName:  "sheets-bucket-name",
			ClixBucketName:    "clix-bucket-name",
			TracksBucketName:  "tracks-bucket-name",
			BackupsBucketName: "backups-bucket-name",
			RedisNamespace:    "redis-namespace",
		},
		TracingConfig: tracing.Config{
			HoneycombWriteKey: "tracing-honeycomb-write-key",
			HoneycombDataset:  "tracing-honeycomb-dataset",
			ServiceName:       "tracing-service-name",
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
