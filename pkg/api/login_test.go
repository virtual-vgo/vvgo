package api

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"golang.org/x/net/publicsuffix"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestLoginHandler_ServeHTTP(t *testing.T) {
	loginHandler := DiscordLoginHandler{
		GuildID:        "test-guild-id",
		RoleVVGOMember: "vvgo-member",
	}
	ts := httptest.NewServer(&loginHandler)
	defer ts.Close()

	discordOAuthTokenJSON := []byte(`{
		"access_token":  "6qrZcUqja7812RVdnEKjpzOL4CvHBFG",
		"token_type":    "Bearer",
		"expires_in":    604800,
		"refresh_token": "D43f5y0ahjqew82jZ4NViEr2YafMKhue",
		"scope":         "identify"
	}`)

	var newServer = func(pre func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if pre != nil {
				pre(w, r)
			}
			switch r.URL.Path {
			case "/users/@me":
				w.Write([]byte(`{"id": "80351110224678912"}`))
			case "/guilds/test-guild-id/members/80351110224678912":
				w.Write([]byte(`{"roles": ["jelly", "donut", "vvgo-member"]}`))
			}
		}))
	}

	t.Run("success", func(t *testing.T) {
		discordTs := newServer(nil)
		defer discordTs.Close()

		loginHandler.Sessions = newSessions("")
		loginHandler.Discord = discord.NewClient(discord.Config{Endpoint: discordTs.URL})
		postAndAssertSession(t, ts.URL, bytes.NewReader(discordOAuthTokenJSON), loginHandler.Sessions)
	})

	t.Run("bad token", func(t *testing.T) {
		discordTs := newServer(nil)
		defer discordTs.Close()

		loginHandler.Sessions = newSessions("")
		loginHandler.Discord = discord.NewClient(discord.Config{Endpoint: discordTs.URL})

		postAndAssertUnauthorized(t, ts.URL, strings.NewReader(`{"access_token":  "6qrZcUqja78}`))
	})

	t.Run("discord identity fails", func(t *testing.T) {
		discordTs := newServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/users/@me" {
				w.WriteHeader(http.StatusBadRequest)
			}
		})
		defer discordTs.Close()

		loginHandler.Sessions = newSessions("")
		loginHandler.Discord = discord.NewClient(discord.Config{Endpoint: discordTs.URL})
		postAndAssertUnauthorized(t, ts.URL, bytes.NewReader(discordOAuthTokenJSON))
	})

	t.Run("discord guild fails", func(t *testing.T) {
		discordTs := newServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/guilds/test-guild-id/members/80351110224678912" {
				w.WriteHeader(http.StatusBadRequest)
			}
		})
		defer discordTs.Close()

		loginHandler.Sessions = newSessions("")
		loginHandler.Discord = discord.NewClient(discord.Config{Endpoint: discordTs.URL})
		postAndAssertUnauthorized(t, ts.URL, bytes.NewReader(discordOAuthTokenJSON))
	})

	t.Run("not a member", func(t *testing.T) {
		discordTs := newServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/guilds/test-guild-id/members/80351110224678912" {
				w.Write([]byte(`{"roles": ["jelly", "donut"]}`))
			}
		})
		defer discordTs.Close()

		loginHandler.Sessions = newSessions("")
		loginHandler.Discord = discord.NewClient(discord.Config{Endpoint: discordTs.URL})
		postAndAssertUnauthorized(t, ts.URL, bytes.NewReader(discordOAuthTokenJSON))
	})
}

func postAndAssertSession(t *testing.T, tsURL string, body io.Reader, sessions *login.Store) {
	ctx := context.Background()
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	require.NoError(t, err, "cookiejar.New")
	client := noFollow(&http.Client{Jar: jar})

	// do the request
	resp, err := client.Post(tsURL, "application/json", body)
	require.NoError(t, err, "client.Post")
	assert.Equal(t, http.StatusFound, resp.StatusCode)

	// check that we get a cookie
	parsed, err := url.Parse(tsURL)
	require.NoError(t, err)
	cookies := jar.Cookies(parsed)
	require.Equal(t, 1, len(cookies), "len(cookies)")
	assert.Equal(t, "vvgo-test-cookie", cookies[0].Name, "cookie name")

	// check that a session exists for the cookie
	var dest login.Identity
	assert.NoError(t, sessions.GetSession(ctx, cookies[0].Value, &dest))
	assert.Equal(t, login.KindDiscord, dest.Kind, "identity.Kind")
	assert.Equal(t, []login.Role{login.RoleVVGOMember}, dest.Roles, "identity.Roles")
}

func postAndAssertUnauthorized(t *testing.T, tsURL string, body io.Reader) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	require.NoError(t, err, "cookiejar.New")
	client := noFollow(&http.Client{Jar: jar})

	// do the request
	resp, err := client.Post(tsURL, "application/json", body)
	require.NoError(t, err, "client.Post")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// check that we get no cookies
	parsed, err := url.Parse(tsURL)
	require.NoError(t, err)
	cookies := jar.Cookies(parsed)
	assert.Equal(t, 0, len(cookies), "len(cookies)")
}

func noFollow(client *http.Client) *http.Client {
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return client
}
