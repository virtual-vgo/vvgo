package api

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/login"
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
		loginHandler.Sessions = newSessions(strings.TrimPrefix(ts.URL, "http://"))

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
		loginHandler.Sessions = newSessions(strings.TrimPrefix(ts.URL, "http://"))

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
	logoutHandler.Sessions = newSessions(strings.TrimPrefix(ts.URL, "http://"))

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

func noFollow(client *http.Client) *http.Client {
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return client
}
