package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"testing"
)

func TestConfig_ParseFile(t *testing.T) {
	want := Config{
		Api: api.ServerConfig{
			ListenAddress:         "listen-address",
			MemberUser:            "member-user",
			MemberPass:            "member-pass",
			DistroBucketName:      "distro-bucket-name",
			RedisNamespace:        "redis-namespace",
			PartsSpreadsheetID:    "parts-spreadsheet-id",
			PartsReadRange:        "parts-read-range",
			DiscordGuildID:        "discord-guild-id",
			DiscordRoleVVGOMember: "discord-role-vvgo-member",
			Login: login.Config{
				CookieName:   "login-cookie-name",
				CookieDomain: "login-cookie-domain",
				CookiePath:   "login-cookie-path",
			},
		},
		Redis: redis.Config{
			Network:  "redis-network",
			Address:  "redis-address",
			PoolSize: 17,
		},
		Minio: storage.Config{
			Endpoint:  "minio-endpoint",
			Region:    "minio-region",
			AccessKey: "minio-access-key",
			SecretKey: "minio-secret-key",
			UseSSL:    true,
		},
		Discord: discord.Config{
			Endpoint:          "discord-endpoint",
			BotAuthToken:      "discord-bot-auth-token",
			OAuthClientID:     "discord-oauth-client-id",
			OAuthClientSecret: "discord-oauth-client-secret",
			OAuthRedirectURI:  "discord-oauth-redirect-uri",
		},
	}

	var got Config
	got.ParseFile("testdata/config.json")
	wantJSON, _ := json.MarshalIndent(want, "", "  ")
	fmt.Println("want json:\n", string(wantJSON))
	assert.Equal(t, want, got)
}
