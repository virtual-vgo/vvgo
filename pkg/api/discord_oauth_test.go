package api

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestDiscordOAuthHandler_buildOAuthRequest(t *testing.T) {

}

func TestDiscordOAuthHandler_doOAuthRequest(t *testing.T) {
	ctx := context.Background()
	oauthHandler := &DiscordOAuthHandler{
		Config: DiscordOAuthHandlerConfig{
			OAuthRedirectURI:  "https://localhost/redirect",
			OAuthClientID:     "test-client-id",
			OAuthClientSecret: "test-client-secret",
		},
	}
	wantForm := make(url.Values)
	wantForm.Add("client_id", "test-client-id")
	wantForm.Add("client_secret", "test-client-secret")
	wantForm.Add("grant_type", "authorization_code")
	wantForm.Add("code", "0xff")
	wantForm.Add("redirect_uri", "https://localhost/redirect")
	wantForm.Add("scope", "identify")
	wantMethod := http.MethodPost
	wantContentType := "application/x-www-form-urlencoded"
	wantUrl := "https://discordapp.com/api/v6/oauth2/token"

	gotRequest, err := oauthHandler.buildOAuthRequest(ctx, "0xff")
	assert.NoError(t, err)
	assert.Equal(t, wantMethod, gotRequest.Method)
	assert.Equal(t, wantUrl, gotRequest.URL.String())
	assert.Equal(t, wantContentType, gotRequest.Header.Get("Content-Type"))
	var buf bytes.Buffer
	buf.ReadFrom(gotRequest.Body)
	assert.Equal(t, wantForm.Encode(), buf.String())
}
