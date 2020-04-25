package api

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestDiscordOAuthHandler_ServeHTTP(t *testing.T) {
	handler := DiscordOAuthHandler{
		Config: DiscordOAuthHandlerConfig{
			BotAuthToken:      "test-bot-auth-token",
			GuildID:           "test-guild-id",
			RoleVVGOMember:    "test-role-vvgo-member",
			OAuthClientID:     "test-oauth-client-id",
			OAuthClientSecret: "test-oauth-client-secret",
			OAuthRedirectURI:  "https://localhost/test-oauth-redirect-uri",
		},
	}
}

