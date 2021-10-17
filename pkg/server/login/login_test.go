package login

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers/test_helpers"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestLoginHandler_ServeHTTP(t *testing.T) {
	ctx := context.Background()

	// password is vvgo-pass
	config.Config.VVGO.MemberPasswordHash = `$2a$10$7FR7RLJNkr1PQV7ahsoPPOV.9orLsENrXi8wnz2mQf8oyKmpnlt2O`

	t.Run("post/failure", func(t *testing.T) {
		urlValues := make(url.Values)
		urlValues.Add("user", "vvgo-member")
		urlValues.Add("pass", "the-wrong-password")
		recorder := httptest.NewRecorder()
		Password(recorder, httptest.NewRequest(http.MethodPost, "/?"+urlValues.Encode(), nil))
		test_helpers.AssertEqualResponse(t, models.ApiResponse{
			Status: models.StatusError,
			Error: &models.ErrorResponse{
				Code:  http.StatusUnauthorized,
				Error: "unauthorized",
			},
		}, recorder.Result())
	})

	t.Run("success", func(t *testing.T) {
		urlValues := make(url.Values)
		urlValues.Add("user", "vvgo-member")
		urlValues.Add("pass", "vvgo-pass")
		recorder := httptest.NewRecorder()
		Password(recorder, httptest.NewRequest(http.MethodPost, "/?"+urlValues.Encode(), nil))

		resp := recorder.Result()
		assert.Equal(t, http.StatusFound, resp.StatusCode)

		// check that we get a cookie
		cookies := resp.Cookies()
		require.Equal(t, 1, len(resp.Cookies()), "len(cookies)")
		assert.Equal(t, "vvgo-sessions", cookies[0].Name, "cookie name")

		// check that a session exists for the cookie
		var dest models.Identity
		assert.NoError(t, GetSession(ctx, cookies[0].Value, &dest))
		assert.Equal(t, models.KindPassword, dest.Kind, "identity.Kind")
		assert.Equal(t, []models.Role{models.RoleVVGOMember}, dest.Roles, "identity.Roles")
	})
}

func TestLogoutHandler_ServeHTTP(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(Logout))
	defer ts.Close()

	// create a session and cookie
	cookie, err := NewCookie(ctx, &models.Identity{
		Kind:  models.KindPassword,
		Roles: []models.Role{"Cheese"},
	}, 3600*time.Second)
	require.NoError(t, err)

	// make the request
	client := http_wrappers.NoFollow(http.DefaultClient)
	req, _ := http.NewRequest(http.MethodGet, ts.URL, nil)
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	require.NoError(t, err, "client.Do")
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/", resp.Header.Get("Location"), "location")

	// check that the session doesn't exist
	var dest models.Identity
	assert.Equal(t, ErrSessionNotFound, GetSession(ctx, cookie.Value, &dest))
}

func TestDiscordOAuthPre_ServeHTTP(t *testing.T) {
	ctx := context.Background()
	ts := httptest.NewServer(http.HandlerFunc(Discord))
	defer ts.Close()

	// make the request
	resp, err := http_wrappers.NoFollow(&http.Client{}).Get(ts.URL)
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
		return ts
	}

	newVVGOServer := func(discordURL string) *httptest.Server {
		config.Config.Discord.Endpoint = discordURL
		config.Config.Discord.BotAuthenticationToken = "test-bot-auth-token"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Discord(w, r.WithContext(ctx))
		}))
		return ts
	}

	doRequest := func(t *testing.T, vvgoURL string, code string, state string, value string) *http.Response {
		req, err := http.NewRequest(http.MethodPost, vvgoURL+"?state="+state+"&code="+code, nil)
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
		resp, err := http_wrappers.NoFollow(&http.Client{}).Do(req)
		require.NoError(t, err, "client.Post")
		return resp
	}

	t.Run("success", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()
		ts := newVVGOServer(discordTs.URL)
		defer ts.Close()

		resp := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusFound, resp.StatusCode)

		// check that we get a cookie
		cookies := resp.Cookies()
		require.Equal(t, 1, len(cookies), "len(cookies)")
		assert.Equal(t, "vvgo-sessions", cookies[0].Name, "cookie name")

		// check that a session exists for the cookie
		var dest models.Identity
		assert.NoError(t, GetSession(context.Background(), cookies[0].Value, &dest))
		assert.Equal(t, models.KindDiscord, dest.Kind, "identity.Kind")
		assert.Equal(t, []models.Role{models.RoleVVGOMember}, dest.Roles, "identity.Roles")
	})

	t.Run("no state", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()
		ts := newVVGOServer(discordTs.URL)
		defer ts.Close()

		// make the request
		resp, err := http_wrappers.NoFollow(&http.Client{}).Get(ts.URL)
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
		ts := newVVGOServer(discordTs.URL)
		defer ts.Close()

		resp := doRequest(t, ts.URL, "fresh", oauthState, oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Empty(t, len(resp.Cookies()), "cookies")
	})

	t.Run("invalid state", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()
		ts := newVVGOServer(discordTs.URL)
		defer ts.Close()
		resp := doRequest(t, ts.URL, oauthCode, "cheese", oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Empty(t, len(resp.Cookies()), "cookies")
	})

	t.Run("invalid value", func(t *testing.T) {
		discordTs := newDiscordServer(nil)
		defer discordTs.Close()
		ts := newVVGOServer(discordTs.URL)
		defer ts.Close()
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
		ts := newVVGOServer(discordTs.URL)
		defer ts.Close()
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
		ts := newVVGOServer(discordTs.URL)
		defer ts.Close()
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
		ts := newVVGOServer(discordTs.URL)
		defer ts.Close()
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
		ts := newVVGOServer(discordTs.URL)
		defer ts.Close()
		resp := doRequest(t, ts.URL, oauthCode, oauthState, oauthValue)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		assert.Empty(t, len(resp.Cookies()), "cookies")
	})
}
