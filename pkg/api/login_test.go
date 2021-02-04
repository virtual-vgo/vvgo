package api

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestLoginView_ServeHTTP(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		server := LoginView{}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		server.ServeHTTP(recorder, request)
		gotResp := recorder.Result()
		assert.Equal(t, http.StatusOK, gotResp.StatusCode)
	})

	t.Run("logged in", func(t *testing.T) {
		ctx := context.Background()
		loginView := LoginView{}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loginView.ServeHTTP(w, r.Clone(context.WithValue(ctx, CtxKeyVVGOIdentity, &login.Identity{Roles: []login.Role{login.RoleVVGOMember}})))
		}))
		defer ts.Close()

		cookie, err := login.NewStore(ctx).NewCookie(ctx, &login.Identity{Roles: []login.Role{login.RoleVVGOMember}}, 600*time.Second)
		require.NoError(t, err, "sessions.NewCookie()")

		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		require.NoError(t, err, "http.NewRequest()")
		req.AddCookie(cookie)
		resp, err := noFollow(nil).Do(req)
		require.NoError(t, err, "http.Do()")
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/login/success", resp.Header.Get("Location"))
	})
}

func TestLoginHandler_ServeHTTP(t *testing.T) {
	ctx := context.Background()
	loginHandler := PasswordLoginHandler{}

	require.NoError(t,
		parse_config.WriteToRedisHash(ctx, "password_login", map[string]string{"vvgo-user": "vvgo-pass"}),
		"redis.Do() failed")

	t.Run("post/failure", func(t *testing.T) {
		ts := httptest.NewServer(&loginHandler)
		defer ts.Close()

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

		client := noFollow(&http.Client{})

		urlValues := make(url.Values)
		urlValues.Add("user", "vvgo-user")
		urlValues.Add("pass", "vvgo-pass")

		// do the request
		resp, err := client.PostForm(ts.URL, urlValues)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// check that we get a cookie
		cookies := resp.Cookies()
		require.Equal(t, 1, len(resp.Cookies()), "len(cookies)")
		assert.Equal(t, "vvgo-sessions", cookies[0].Name, "cookie name")

		// check that a session exists for the cookie
		var dest login.Identity
		assert.NoError(t, login.NewStore(ctx).GetSession(ctx, cookies[0].Value, &dest))
		assert.Equal(t, login.KindPassword, dest.Kind, "identity.Kind")
		assert.Equal(t, []login.Role{login.RoleVVGOMember}, dest.Roles, "identity.Roles")
	})
}

func TestLogoutHandler_ServeHTTP(t *testing.T) {
	ctx := context.Background()
	logoutHandler := LogoutHandler{}

	ts := httptest.NewServer(&logoutHandler)
	defer ts.Close()

	// create a session and cookie
	cookie, err := login.NewStore(ctx).NewCookie(ctx, &login.Identity{
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
	assert.Equal(t, login.ErrSessionNotFound, login.NewStore(ctx).GetSession(ctx, cookie.Value, &dest))
}

func TestDiscordOAuthPre_ServeHTTP(t *testing.T) {

	ctx := context.Background()
	handler := DiscordLoginHandler{}
	ts := httptest.NewServer(&handler)
	defer ts.Close()

	// make the request
	resp, err := noFollow(&http.Client{}).Get(ts.URL)
	require.NoError(t, err, "client.Get()")
	require.Equal(t, http.StatusFound, resp.StatusCode, "status code")

	// parse the location url
	location, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err, "url.Parse()")
	query := location.Query()
	assert.Equal(t, "/api/oauth2/authorize", location.Path)

	// parse the state and value
	cookies := resp.Cookies()
	require.NotEmpty(t, cookies, "cookies")
	oauthState := query.Get("state")
	assert.NotEmpty(t, oauthState, "oauth state")
	oauthValue := cookies[0].Value
	var wantValue string
	err = redis.Do(ctx, redis.Cmd(&wantValue, "GET", "oauth_state:"+oauthState))
	require.NoError(t, err, "redis.Do()")
	assert.Equal(t, wantValue, oauthValue, "cookie value")
}

func TestDiscordLoginHandler_ServeHTTP(t *testing.T) {
	ctx := context.Background()
	oauthState := "test-oauth-state"
	oauthValue := "test-oauth-value"
	oauthCode := "test-oauth-code"
	require.NoError(t, redis.Do(ctx, redis.Cmd(nil, "SETEX", "oauth_state:"+oauthState, "300", oauthValue)))

	loginHandler := DiscordLoginHandler{}
	ts := httptest.NewServer(&loginHandler)
	defer ts.Close()

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
			case "/guilds/" + discord.VVGOGuildID + "/members/80351110224678912":
				if r.Header.Get("Authorization") == "Bot test-bot-auth-token" {
					w.Write([]byte(`{"roles": ["jelly", "donut", "` + discord.VVGOVerifiedMemberRoleID + `"]}`))
				} else {
					http.Error(w, "access denied; invalid authorization: "+r.Header.Get("Authorization"), http.StatusUnauthorized)
				}
			}
		}))

		discordConfig := discord.Config{
			Endpoint:               ts.URL,
			BotAuthenticationToken: "test-bot-auth-token",
		}
		require.NoError(t, parse_config.WriteToRedisHash(ctx, "discord", &discordConfig), "redis.Do() failed")
		return ts
	}

	doRequest := func(t *testing.T, tsURL string, code string, state string, value string) *http.Response {
		req, err := http.NewRequest(http.MethodPost, tsURL+"?state="+state+"&code="+code, nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{
			Name:     CookieOAuthState,
			Value:    value,
			Path:     "/login/discord",
			Domain:   "",
			Expires:  time.Now().Add(300 * time.Second),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		resp, err := noFollow(&http.Client{}).Do(req)
		require.NoError(t, err, "client.Post")
		return resp
	}

	t.Run("success", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()

		resp := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// check that we get a cookie
		cookies := resp.Cookies()
		require.Equal(t, 1, len(cookies), "len(cookies)")
		assert.Equal(t, "vvgo-sessions", cookies[0].Name, "cookie name")

		// check that a session exists for the cookie
		var dest login.Identity
		assert.NoError(t, login.NewStore(ctx).GetSession(context.Background(), cookies[0].Value, &dest))
		assert.Equal(t, login.KindDiscord, dest.Kind, "identity.Kind")
		assert.Equal(t, []login.Role{login.RoleVVGOMember}, dest.Roles, "identity.Roles")
	})

	t.Run("no state", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()

		// make the request
		resp, err := noFollow(&http.Client{}).Get(ts.URL)
		require.NoError(t, err, "client.Get()")
		require.Equal(t, http.StatusFound, resp.StatusCode, "status code")

		// parse the location url
		location, err := url.Parse(resp.Header.Get("Location"))
		require.NoError(t, err, "url.Parse()")
		query := location.Query()

		assert.Equal(t, "discord.com", location.Host)
		assert.NotEmpty(t, query.Get("state"), "state")

		// parse the state and value
		cookies := resp.Cookies()
		require.NotEmpty(t, cookies, "cookies")
		oauthState := query.Get("state")
		oauthValue := cookies[0].Value
		var wantValue string
		err = redis.Do(ctx, redis.Cmd(&wantValue, "GET", "oauth_state:"+oauthState))
		require.NoError(t, err, "redis.Do()")
		assert.Equal(t, wantValue, oauthValue, "cookie value")
	})

	t.Run("invalid code", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()
		resp := doRequest(t, ts.URL, "fresh", oauthState, oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Empty(t, len(resp.Cookies()), "cookies")
	})

	t.Run("invalid state", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()
		resp := doRequest(t, ts.URL, oauthCode, "cheese", oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Empty(t, len(resp.Cookies()), "cookies")
	})

	t.Run("invalid value", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()
		resp := doRequest(t, ts.URL, oauthCode, oauthState, "danish")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Empty(t, len(resp.Cookies()), "cookies")
	})

	t.Run("bad token", func(t *testing.T) {
		discordTs := newDiscordServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/oauth2/token" {
				w.Write([]byte(`{}`))
			}
		})
		defer discordTs.Close()
		resp := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Empty(t, len(resp.Cookies()), "cookies")
	})

	t.Run("discord identity fails", func(t *testing.T) {
		discordTs := newDiscordServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/users/@me" {
				w.WriteHeader(http.StatusBadRequest)
			}
		})
		defer discordTs.Close()
		resp := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Empty(t, len(resp.Cookies()), "cookies")
	})

	t.Run("discord guild fails", func(t *testing.T) {
		discordTs := newDiscordServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/guilds/"+discord.VVGOGuildID+"/members/80351110224678912" {
				w.WriteHeader(http.StatusBadRequest)
			}
		})
		defer discordTs.Close()
		resp := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Empty(t, len(resp.Cookies()), "cookies")
	})

	t.Run("not a member", func(t *testing.T) {
		discordTs := newDiscordServer(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/guilds/"+discord.VVGOGuildID+"/members/80351110224678912" {
				w.Write([]byte(`{"roles": ["jelly", "donut"]}`))
			}
		})
		defer discordTs.Close()
		resp := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Empty(t, len(resp.Cookies()), "cookies")
	})
}
