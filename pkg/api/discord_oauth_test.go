package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestDiscordOAuthHandler_buildOAuthRequest(t *testing.T) {
	ctx := context.Background()
	oauthHandler := &DiscordOAuthHandler{
		Config: DiscordOAuthHandlerConfig{
			OAuthRedirectURI:  "https://localhost/redirect",
			OAuthClientID:     "test-client-id",
			OAuthClientSecret: "test-client-secret",
		},
	}
	wantForm := make(url.Values)
	wantForm.Add("grant_type", "authorization_code")
	wantForm.Add("code", "this is a code")
	wantForm.Add("redirect_uri", "https://localhost/redirect")
	wantForm.Add("scope", "identify")
	wantMethod := http.MethodPost
	wantUrl := "https://discordapp.com/api/v6/oauth2/token"
	wantUser := "test-client-id"
	wantPass := "test-client-secret"

	gotRequest, err := oauthHandler.buildOAuthRequest(ctx, "this is a code")
	assert.NoError(t, err)
	assert.Equal(t, wantMethod, gotRequest.Method)
	assert.Equal(t, wantUrl, gotRequest.URL.String())
	gotUser, gotPass, _ := gotRequest.BasicAuth()
	assert.Equal(t, wantUser, gotUser)
	assert.Equal(t, wantPass, gotPass)
}
