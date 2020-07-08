package discord

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

func TestClient_QueryOAuth(t *testing.T) {
	ctx := context.Background()
	client := Client{
		config: Config{
			BotAuthToken:      "test-bot-auth-token",
			OAuthClientID:     "test-oauth-client-id",
			OAuthClientSecret: "test-oauth-client-secret",
			OAuthRedirectURI:  "https://localhost/test-oauth-redirect-uri",
		},
	}

	var gotRequest *http.Request
	var gotForm string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r

		var buf bytes.Buffer
		_, err := buf.ReadFrom(gotRequest.Body)
		require.NoError(t, err)
		gotForm = buf.String()

		// https://discordapp.com/developers/docs/topics/oauth2#authorization-code-grant-access-token-response
		w.Write([]byte(`{
			"access_token": "6qrZcUqja7812RVdnEKjpzOL4CvHBFG",
			"token_type": "Bearer",
			"expires_in": 604800,
			"refresh_token": "D43f5y0ahjqew82jZ4NViEr2YafMKhue",
			"scope": "identify"
		}`))
	}))
	defer ts.Close()
	client.config.Endpoint = ts.URL
	gotToken, gotError := client.QueryOAuth(ctx, "test-code")
	require.NoError(t, gotError)
	assert.Equal(t, http.MethodPost, gotRequest.Method)
	assert.Equal(t, "/oauth2/token", gotRequest.URL.String())
	assert.Equal(t, "application/x-www-form-urlencoded", gotRequest.Header.Get("Content-Type"))

	wantForm := make(url.Values)
	wantForm.Add("client_id", "test-oauth-client-id")
	wantForm.Add("client_secret", "test-oauth-client-secret")
	wantForm.Add("grant_type", "authorization_code")
	wantForm.Add("code", "test-code")
	wantForm.Add("redirect_uri", "https://localhost/test-oauth-redirect-uri")
	wantForm.Add("scope", "identify")
	assert.Equal(t, wantForm.Encode(), gotForm)

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
	client := Client{
		config: Config{
			BotAuthToken: "test-bot-auth-token",
		},
	}
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
		w.Write([]byte(`{
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
	client.config.Endpoint = ts.URL
	gotUser, gotError := client.QueryIdentity(ctx, token)
	require.NoError(t, gotError)
	assert.Equal(t, http.MethodGet, gotRequest.Method)
	assert.Equal(t, "/users/@me", gotRequest.URL.String())
	assert.Equal(t, []string{"Bearer 6qrZcUqja7812RVdnEKjpzOL4CvHBFG"}, gotRequest.Header["Authorization"])
	assert.Equal(t, &User{ID: "80351110224678912"}, gotUser)
}

func TestClient_QueryGuildMember(t *testing.T) {
	ctx := context.Background()
	client := Client{
		config: Config{
			BotAuthToken: "test-bot-auth-token",
		},
	}

	var gotRequest *http.Request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r
		// https://discordapp.com/developers/docs/resources/guild#guild-member-object-example-guild-member
		w.Write([]byte(`{
			"user": {},
			"nick": "NOT API SUPPORT",
			"roles": ["jelly", "donut"],
			"joined_at": "2015-04-26T06:26:56.936000+00:00",
			"deaf": false,
			"mute": false
		}`))
	}))
	defer ts.Close()
	client.config.Endpoint = ts.URL
	gotMember, gotError := client.QueryGuildMember(ctx, "test-guild-id", "test-user-id")
	require.NoError(t, gotError)
	assert.Equal(t, http.MethodGet, gotRequest.Method)
	assert.Equal(t, "/guilds/test-guild-id/members/test-user-id", gotRequest.URL.String())
	assert.Equal(t, []string{"Bot test-bot-auth-token"}, gotRequest.Header["Authorization"])
	assert.Equal(t, &GuildMember{Roles: []string{"jelly", "donut"}}, gotMember)
}
