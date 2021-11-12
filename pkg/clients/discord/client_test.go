package discord

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestClient_QueryOAuth(t *testing.T) {
	ctx := context.Background()
	config.Env.Discord.BotAuthenticationToken = "test-bot-auth-token"
	config.Env.Discord.OAuthClientSecret = "test-oauth-client-secret"

	var gotRequest *http.Request
	var gotForm string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r

		var buf bytes.Buffer
		_, err := buf.ReadFrom(gotRequest.Body)
		require.NoError(t, err)
		gotForm = buf.String()

		// https://discordapp.com/developers/docs/topics/oauth2#authorization-code-grant-access-token-response
		_, _ = w.Write([]byte(`{
			"access_token": "6qrZcUqja7812RVdnEKjpzOL4CvHBFG",
			"token_type": "Bearer",
			"expires_in": 604800,
			"refresh_token": "D43f5y0ahjqew82jZ4NViEr2YafMKhue",
			"scope": "identify"
		}`))
	}))
	defer ts.Close()
	config.Env.Discord.Endpoint = ts.URL
	gotToken, gotError := GetOAuthToken(ctx, "test-code")
	require.NoError(t, gotError)
	assert.Equal(t, http.MethodPost, gotRequest.Method)
	assert.Equal(t, "/oauth2/token", gotRequest.URL.String())
	assert.Equal(t, "application/x-www-form-urlencoded", gotRequest.Header.Get("Content-Type"))

	wantForm := make(url.Values)
	wantForm.Add("client_id", OAuthClientID)
	wantForm.Add("client_secret", "test-oauth-client-secret")
	wantForm.Add("grant_type", "authorization_code")
	wantForm.Add("code", "test-code")
	wantForm.Add("redirect_uri", "https://vvgo.org/login/discord")
	wantForm.Add("scope", "identify")
	assert.Equal(t, wantForm.Encode(), gotForm)

	//goland:noinspection SpellCheckingInspection
	assert.Equal(t, &OAuthToken{
		AccessToken:  "6qrZcUqja7812RVdnEKjpzOL4CvHBFG",
		TokenType:    "Bearer",
		ExpiresIn:    604800,
		RefreshToken: "D43f5y0ahjqew82jZ4NViEr2YafMKhue",
		Scope:        "identify",
	}, gotToken)
}

func TestClient_QueryIdentity(t *testing.T) {
	ctx := context.Background()
	config.Env.Discord.BotAuthenticationToken = "test-bot-auth-token"
	token := &OAuthToken{
		AccessToken:  "6qrZcUqja7812RVdnEKjpzOL4CvHBFG",
		TokenType:    "Bearer",
		ExpiresIn:    604800,
		RefreshToken: "D43f5y0ahjqew82jZ4NViEr2YafMKhue",
		Scope:        "identify",
	}

	var gotRequest *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r
		// https://discordapp.com/developers/docs/resources/user#user-object-example-user
		_, _ = w.Write([]byte(`{
			"id": "80351110224678912",
			"username": "Nelly",
			"discriminator": "1337",
			"avatar": "8342729096ea3675442027381ff50dfe",
			"verified": true,
			"email": "nelly@discordapp.com",
			"flags": 64,
			"premium_type": 1,
			"public_flags": 64
		}`))
	}))
	defer ts.Close()
	config.Env.Discord.Endpoint = ts.URL
	gotUser, gotError := GetIdentity(ctx, token)
	require.NoError(t, gotError)
	assert.Equal(t, http.MethodGet, gotRequest.Method)
	assert.Equal(t, "/users/@me", gotRequest.URL.String())
	//goland:noinspection SpellCheckingInspection
	assert.Equal(t, []string{"Bearer 6qrZcUqja7812RVdnEKjpzOL4CvHBFG"}, gotRequest.Header["Authorization"])
	assert.Equal(t, &User{ID: "80351110224678912", Username: "Nelly"}, gotUser)
}

func TestClient_QueryGuildMember(t *testing.T) {
	ctx := context.Background()
	config.Env.Discord.BotAuthenticationToken = "test-bot-auth-token"

	var gotRequest *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r
		// https://discordapp.com/developers/docs/resources/guild#guild-member-object-example-guild-member
		_, _ = w.Write([]byte(`{
			"user": {},
			"nick": "NOT API SUPPORT",
			"roles": ["jelly", "donut"],
			"joined_at": "2015-04-26T06:26:56.936000+00:00",
			"deaf": false,
			"mute": false
		}`))
	}))
	defer ts.Close()
	config.Env.Discord.Endpoint = ts.URL
	gotMember, gotError := GetGuildMember(ctx, "test-user-id")
	require.NoError(t, gotError)
	assert.Equal(t, http.MethodGet, gotRequest.Method)
	assert.Equal(t, "/guilds/690626216637497425/members/test-user-id", gotRequest.URL.String())
	assert.Equal(t, []string{"Bot test-bot-auth-token"}, gotRequest.Header["Authorization"])
	assert.Equal(t, &GuildMember{Nick: "NOT API SUPPORT", Roles: []string{"jelly", "donut"}}, gotMember)
}
