package main

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/facebook"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"os"
	"testing"
)

func TestConfig_ParseEnv(t *testing.T) {
	envs := map[string]string{
		"API_LISTEN_ADDRESS":           "listen-address",
		"API_DISTRO_BUCKET_NAME":       "distro-bucket-name",
		"API_BACKUPS_BUCKET_NAME":      "backups-bucket-name",
		"API_REDIS_NAMESPACE":          "redis-namespace",
		"API_MEMBER_USER":              "member-user",
		"API_MEMBER_PASS":              "member-pass",
		"REDIS_ADDRESS":                "redis-address",
		"REDIS_NETWORK":                "redis-network",
		"REDIS_POOL_SIZE":              "17",
		"MINIO_ENDPOINT":               "minio-endpoint",
		"MINIO_REGION":                 "minio-region",
		"MINIO_ACCESSKEY":              "minio-access-key",
		"MINIO_SECRETKEY":              "minio-secret-key",
		"MINIO_USESSL":                 "true",
		"API_UPLOADER_TOKEN":           "uploader-token",
		"API_DEVELOPER_TOKEN":          "developer-token",
		"API_LOGIN_COOKIE_NAME":        "login-cookie-name",
		"API_LOGIN_COOKIE_DOMAIN":      "login-cookie-domain",
		"API_LOGIN_COOKIE_PATH":        "login-cookie-path",
		"API_DISCORD_GUILD_ID":         "discord-guild-id",
		"API_FACEBOOK_GROUP_ID":        "facebook-group-id",
		"API_DISCORD_ROLE_VVGO_MEMBER": "discord-role-vvgo-member",
		"API_DISCORD_LOGIN_URL":        "discord-login-url",
		"API_PARTS_SPREADSHEET_ID":     "parts-spreadsheet-id",
		"API_PARTS_READ_RANGE":         "parts-read-range",
		"DISCORD_BOT_AUTH_TOKEN":       "discord-bot-auth-token",
		"DISCORD_ENDPOINT":             "discord-endpoint",
		"DISCORD_OAUTH_CLIENT_ID":      "discord-oauth-client-id",
		"DISCORD_OAUTH_CLIENT_SECRET":  "discord-oauth-client-secret",
		"DISCORD_OAUTH_REDIRECT_URI":   "discord-oauth-redirect-uri",
		"FACEBOOK_BOT_AUTH_TOKEN":      "facebook-bot-auth-token",
		"FACEBOOK_ENDPOINT":            "facebook-endpoint",
		"FACEBOOK_OAUTH_CLIENT_ID":     "facebook-oauth-client-id",
		"FACEBOOK_OAUTH_CLIENT_SECRET": "facebook-oauth-client-secret",
		"FACEBOOK_OAUTH_REDIRECT_URI":  "facebook-oauth-redirect-uri",
	}
	want := Config{
		ApiConfig: api.ServerConfig{
			ListenAddress:         "listen-address",
			MemberUser:            "member-user",
			MemberPass:            "member-pass",
			UploaderToken:         "uploader-token",
			DeveloperToken:        "developer-token",
			DistroBucketName:      "distro-bucket-name",
			BackupsBucketName:     "backups-bucket-name",
			RedisNamespace:        "redis-namespace",
			PartsSpreadsheetID:    "parts-spreadsheet-id",
			PartsReadRange:        "parts-read-range",
			DiscordGuildID:        "discord-guild-id",
			DiscordRoleVVGOMember: "discord-role-vvgo-member",
			DiscordLoginURL:       "discord-login-url",
			FacebookGroupID:       "facebook-group-id",
			Login: login.Config{
				CookieName:   "login-cookie-name",
				CookieDomain: "login-cookie-domain",
				CookiePath:   "login-cookie-path",
			},
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
		DiscordConfig: discord.Config{
			Endpoint:          "discord-endpoint",
			BotAuthToken:      "discord-bot-auth-token",
			OAuthClientID:     "discord-oauth-client-id",
			OAuthClientSecret: "discord-oauth-client-secret",
			OAuthRedirectURI:  "discord-oauth-redirect-uri",
		},
		FacebookConfig: facebook.Config{
			Endpoint:          "facebook-endpoint",
			OAuthClientID:     "facebook-oauth-client-id",
			OAuthClientSecret: "facebook-oauth-client-secret",
			OAuthRedirectURI:  "facebook-oauth-redirect-uri",
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
