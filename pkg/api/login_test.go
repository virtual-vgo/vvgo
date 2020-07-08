package api

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestLoginHandler_ServeHTTP(t *testing.T) {
	loginHandler := PasswordLoginHandler{
		Logins: map[[2]string][]login.Role{
			{"vvgo-user", "vvgo-pass"}: {"vvgo-member"},
		},
	}

	t.Run("post/failure", func(t *testing.T) {
		ts := httptest.NewServer(&loginHandler)
		defer ts.Close()
		loginHandler.Sessions = newSessions()

		urlValues := make(url.Values)
		urlValues.Add("user", "vvgo-user")
		urlValues.Add("pass", "the-wrong-password")
		resp, err := noFollow(http.DefaultClient).PostForm(ts.URL, urlValues)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		var gotBody bytes.Buffer
		gotBody.ReadFrom(resp.Body)
		assert.Equal(t, "authorization failed", strings.TrimSpace(gotBody.String()), "body")
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		ts := httptest.NewServer(&loginHandler)
		defer ts.Close()
		loginHandler.Sessions = newSessions()

		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		require.NoError(t, err, "cookiejar.New")
		client := noFollow(&http.Client{Jar: jar})

		urlValues := make(url.Values)
		urlValues.Add("user", "vvgo-user")
		urlValues.Add("pass", "vvgo-pass")

		// do the request
		resp, err := client.PostForm(ts.URL, urlValues)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, http.StatusFound, resp.StatusCode)

		// check that we get a cookie
		tsURL, err := url.Parse(ts.URL)
		require.NoError(t, err)
		cookies := jar.Cookies(tsURL)
		require.Equal(t, 1, len(cookies), "len(cookies)")
		assert.Equal(t, "vvgo-test-cookie", cookies[0].Name, "cookie name")

		// check that a session exists for the cookie
		var dest login.Identity
		assert.NoError(t, loginHandler.Sessions.GetSession(ctx, cookies[0].Value, &dest))
		assert.Equal(t, login.KindPassword, dest.Kind, "identity.Kind")
		assert.Equal(t, []login.Role{login.RoleVVGOMember}, dest.Roles, "identity.Roles")
	})
}

func TestLogoutHandler_ServeHTTP(t *testing.T) {
	ctx := context.Background()
	logoutHandler := LogoutHandler{}

	ts := httptest.NewServer(&logoutHandler)
	defer ts.Close()
	logoutHandler.Sessions = newSessions()

	// create a session and cookie
	cookie, err := logoutHandler.Sessions.NewCookie(ctx, &login.Identity{
		Kind:  login.KindPassword,
		Roles: []login.Role{"Cheese"},
	}, 3600*time.Second)
	require.NoError(t, err)

	// make the request
	client := noFollow(http.DefaultClient)
	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	require.NoError(t, err, "client.Do")
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/", resp.Header.Get("Location"), "location")

	// check that the session doesn't exist
	var dest login.Identity
	assert.Equal(t, login.ErrSessionNotFound, logoutHandler.Sessions.GetSession(ctx, cookie.Value, &dest))
}

func TestDiscordOAuthPre_ServeHTTP(t *testing.T) {

}

func TestDiscordLoginHandler_ServeHTTP(t *testing.T) {
	oauthNamespace := newNamespace()
	oauthState := "test-oauth-state"
	oauthValue := "test-oauth-value"
	oauthCode := "test-oauth-code"
	require.NoError(t, redis.Do(context.Background(), redis.Cmd(nil, "SETEX", oauthNamespace+":discord_oauth_pre:"+oauthState, "300", oauthValue)))

	loginHandler := DiscordLoginHandler{
		GuildID:        "test-guild-id",
		RoleVVGOMember: "vvgo-member",
		Namespace:      oauthNamespace,
	}
	ts := httptest.NewServer(&loginHandler)
	defer ts.Close()
	parsed, err := url.Parse(ts.URL)
	require.NoError(t, err)

	discordOAuthTokenJSON := []byte(`{
		"access_token":  "6qrZcUqja7812RVdnEKjpzOL4CvHBFG",
		"token_type":    "Bearer",
		"expires_in":    604800,
		"refresh_token": "D43f5y0ahjqew82jZ4NViEr2YafMKhue",
		"scope":         "identify"
	}`)

	newDiscordServer := func(pre func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if pre != nil {
				pre(w, r)
			}
			switch r.URL.Path {
			case "/oauth2/token":
				if r.FormValue("code") == "test-oauth-code" {
					w.Write(discordOAuthTokenJSON)
				} else {
					http.Error(w, "access denied; invalid code: "+r.Form.Get("code"), http.StatusUnauthorized)
				}
			case "/users/@me":
				if r.Header.Get("Authorization") == "Bearer 6qrZcUqja7812RVdnEKjpzOL4CvHBFG" {
					w.Write([]byte(`{"id": "80351110224678912"}`))
				} else {
					http.Error(w, "access denied; invalid authorization: "+r.Header.Get("Authorization"), http.StatusUnauthorized)
				}
			case "/guilds/test-guild-id/members/80351110224678912":
				if r.Header.Get("Authorization") == "Bot test-bot-auth-token" {
					w.Write([]byte(`{"roles": ["jelly", "donut", "vvgo-member"]}`))
				} else {
					http.Error(w, "access denied; invalid authorization: "+r.Header.Get("Authorization"), http.StatusUnauthorized)
				}
			}
		}))
		discord.Initialize(discord.Config{
			Endpoint:     ts.URL,
			BotAuthToken: "test-bot-auth-token",
		})
		return ts
	}

	doRequest := func(t *testing.T, tsURL string, code string, state string, value string) (*http.Response, http.CookieJar) {
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		require.NoError(t, err, "cookiejar.New")
		client := noFollow(&http.Client{Jar: jar})

		req, err := http.NewRequest(http.MethodPost, tsURL+"?state="+state+"&code="+code, nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:     DiscordOAuthPreCookie,
			Value:    value,
			Path:     "/login/discord",
			Domain:   "",
			Expires:  time.Now().Add(300 * time.Second),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		resp, err := client.Do(req)
		require.NoError(t, err, "client.Post")
		return resp, client.Jar
	}

	t.Run("success", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()
		loginHandler.Sessions = newSessions()

		resp, jar := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusFound, resp.StatusCode)

		// check that we get a cookie
		parsed, err := url.Parse(ts.URL)
		require.NoError(t, err)
		cookies := jar.Cookies(parsed)
		require.Equal(t, 1, len(cookies), "len(cookies)")
		assert.Equal(t, "vvgo-test-cookie", cookies[0].Name, "cookie name")

		// check that a session exists for the cookie
		var dest login.Identity
		assert.NoError(t, loginHandler.Sessions.GetSession(context.Background(), cookies[0].Value, &dest))
		assert.Equal(t, login.KindDiscord, dest.Kind, "identity.Kind")
		assert.Equal(t, []login.Role{login.RoleVVGOMember}, dest.Roles, "identity.Roles")
	})

	t.Run("no state", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()
		loginHandler.Sessions = newSessions()
		resp, jar := doRequest(t, ts.URL, oauthCode, "", oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, 0, len(jar.Cookies(parsed)), "len(cookies)")
	})

	t.Run("invalid state", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()
		loginHandler.Sessions = newSessions()
		resp, jar := doRequest(t, ts.URL, oauthCode, "cheese", oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, 0, len(jar.Cookies(parsed)), "len(cookies)")
	})

	t.Run("invalid value", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()
		loginHandler.Sessions = newSessions()
		resp, jar := doRequest(t, ts.URL, oauthCode, oauthState, "danish")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, 0, len(jar.Cookies(parsed)), "len(cookies)")
	})

	t.Run("bad token", func(t *testing.T) {
		discordTs := newDiscordServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/oauth2/token" {
				w.Write([]byte(`{}`))
			}
		})
		defer discordTs.Close()
		loginHandler.Sessions = newSessions()
		resp, jar := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, 0, len(jar.Cookies(parsed)), "len(cookies)")
	})

	t.Run("discord identity fails", func(t *testing.T) {
		discordTs := newDiscordServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/users/@me" {
				w.WriteHeader(http.StatusBadRequest)
			}
		})
		defer discordTs.Close()
		loginHandler.Sessions = newSessions()
		resp, jar := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, 0, len(jar.Cookies(parsed)), "len(cookies)")
	})

	t.Run("discord guild fails", func(t *testing.T) {
		discordTs := newDiscordServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/guilds/test-guild-id/members/80351110224678912" {
				w.WriteHeader(http.StatusBadRequest)
			}
		})
		defer discordTs.Close()
		loginHandler.Sessions = newSessions()
		resp, jar := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, 0, len(jar.Cookies(parsed)), "len(cookies)")
	})

	t.Run("not a member", func(t *testing.T) {
		discordTs := newDiscordServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/guilds/test-guild-id/members/80351110224678912" {
				w.Write([]byte(`{"roles": ["jelly", "donut"]}`))
			}
		})
		defer discordTs.Close()
		loginHandler.Sessions = newSessions()
		resp, jar := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Equal(t, 0, len(jar.Cookies(parsed)), "len(cookies)")
	})
}
