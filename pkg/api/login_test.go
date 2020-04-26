package api

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/access"
	"github.com/virtual-vgo/vvgo/pkg/locker"
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
	secret := access.Secret{1, 2, 3, 4}
	loginHandler := PasswordLoginHandler{
		Sessions: access.NewStore(locker.NewLocksmith(locker.Config{}), access.Config{Secret: secret, CookieName: "vvgo-cookie"}),
		Logins: []PasswordLogin{
			{
				User:  "vvgo-user",
				Pass:  "vvgo-pass",
				Roles: []access.Role{"vvgo-member"},
			},
		},
	}

	t.Run("post/failure", func(t *testing.T) {
		loginHandler.Sessions.Init(context.Background())
		ts := httptest.NewServer(loginHandler)
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
		loginHandler.Sessions.Init(context.Background())
		ts := httptest.NewServer(loginHandler)
		defer ts.Close()

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
		assert.Equal(t, "vvgo-cookie", cookies[0].Name, "cookie name")

		// check that a session exists for the cookie
		var session access.Session
		require.NoError(t, session.DecodeCookie(access.Secret{1, 2, 3, 4}, cookies[0]), "session.DecodeCookie")

		// check that the identity is what we expect
		var identity access.Identity
		require.NoError(t, loginHandler.Sessions.GetIdentity(ctx, session.ID, &identity))
		assert.Equal(t, access.KindPassword, identity.Kind, "identity.Kind")
		assert.Equal(t, []access.Role{access.RoleVVGOMember}, identity.Roles, "identity.Roles")
	})
}

func TestLogoutHandler_ServeHTTP(t *testing.T) {
	ctx := context.Background()
	logoutHandler := LogoutHandler{
		Sessions: access.NewStore(locker.NewLocksmith(locker.Config{}), access.Config{CookieName: "vvgo-cookie"}),
	}

	logoutHandler.Sessions.Init(context.Background())
	ts := httptest.NewServer(logoutHandler)
	defer ts.Close()
	tsUrl, err := url.Parse(ts.URL)
	require.NoError(t, err, "url.Parse()")

	// create a session and cookie
	session := logoutHandler.Sessions.NewSession(time.Now().Add(7 * 24 * 3600 * time.Second))
	cookie := logoutHandler.Sessions.NewCookie(session)
	assert.NoError(t, logoutHandler.Sessions.StoreIdentity(ctx, session.ID, &access.Identity{
		Kind:  access.KindPassword,
		Roles: []access.Role{"cheese"},
	}))

	// set the cookie on the client
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	require.NoError(t, err, "cookiejar.New")
	jar.SetCookies(tsUrl, []*http.Cookie{cookie})

	// make the request
	client := noFollow(&http.Client{Jar: jar})
	resp, err := client.Get(ts.URL)
	require.NoError(t, err, "client.Get")
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "/", resp.Header.Get("Location"), "location")

	// check that the session doesn't exist
	var dest access.Identity
	assert.Equal(t, access.ErrSessionNotFound, logoutHandler.Sessions.GetIdentity(ctx, session.ID, &dest))
}

func noFollow(client *http.Client) *http.Client {
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return client
}
