package api

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/virtual-vgo/vvgo/pkg/access"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestLoginHandler_ServeHTTP(t *testing.T) {
	loginHandler := LoginHandler{
		Sessions: sessions.NewStore(sessions.Secret{}, sessions.Config{CookieName: "vvgo-cookie"}),
		Logins: []Login{
			{
				User:  "vvgo-user",
				Pass:  "vvgo-pass",
				Roles: []access.Role{"vvgo-member"},
			},
		},
	}

	t.Run("get/redirect", func(t *testing.T) {
		ctx := context.Background()
		loginHandler.Sessions.Init(context.Background())
		ts := httptest.NewServer(loginHandler)
		defer ts.Close()
		tsUrl, err := url.Parse(ts.URL)
		require.NoError(t, err, "url.Parse()")

		// create a session and cookie
		session := loginHandler.Sessions.NewSession(time.Now().Add(7 * 24 * 3600 * time.Second))
		cookie := loginHandler.Sessions.NewCookie(session)
		assert.NoError(t, loginHandler.Sessions.StoreIdentity(ctx, session.ID, &sessions.Identity{
			Kind:  sessions.IdentityPassword,
			Roles: []access.Role{"cheese"},
		}))

		// set the cookie on the client
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		require.NoError(t, err, "cookiejar.New")
		jar.SetCookies(tsUrl, []*http.Cookie{cookie})

		client := noFollow(&http.Client{Jar: jar})
		resp, err := client.Get(ts.URL)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/", resp.Header.Get("Location"), "location")
	})

	t.Run("get/view", func(t *testing.T) {
		wantCode := http.StatusOK
		wantBytes, err := ioutil.ReadFile("testdata/login.html")
		if err != nil {
			t.Fatalf("ioutil.ReadFile() failed: %v", err)
		}

		loginHandler.Sessions.Init(context.Background())
		ts := httptest.NewServer(loginHandler)
		defer ts.Close()

		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		require.NoError(t, err, "cookiejar.New")
		client := noFollow(&http.Client{Jar: jar})
		resp, err := client.Get(ts.URL)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, wantCode, resp.StatusCode)
		var respBody bytes.Buffer
		_, err = respBody.ReadFrom(resp.Body)
		require.NoError(t, err, "resp.Body.Read() failed")
		origBody := strings.TrimSpace(respBody.String())

		m := minify.New()
		m.AddFunc("text/html", html.Minify)
		var gotBuf bytes.Buffer
		if err := m.Minify("text/html", &gotBuf, &respBody); err != nil {
			panic(err)
		}
		gotBody := gotBuf.String()

		var wantBuf bytes.Buffer
		if err := m.Minify("text/html", &wantBuf, bytes.NewReader(wantBytes)); err != nil {
			panic(err)
		}
		wantBody := wantBuf.String()
		if !assert.Equal(t, wantBody, gotBody, "body") {
			t.Logf("Got Body:\n%s\n", origBody)
		}
	})

	t.Run("post/failure", func(t *testing.T) {
		loginHandler.Sessions.Init(context.Background())
		ts := httptest.NewServer(loginHandler)
		defer ts.Close()

		urlValues := make(url.Values)
		urlValues.Add("user", "vvgo-user")
		urlValues.Add("pass", "vvgo-user")
		resp, err := noFollow(http.DefaultClient).PostForm(ts.URL, urlValues)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		var gotBody bytes.Buffer
		gotBody.ReadFrom(resp.Body)
		assert.Equal(t, "authorization failed", strings.TrimSpace(gotBody.String()), "body")
	})

	t.Run("post/success", func(t *testing.T) {
		loginHandler.Sessions.Init(context.Background())
		ts := httptest.NewServer(loginHandler)
		defer ts.Close()

		urlValues := make(url.Values)
		urlValues.Add("user", "vvgo-user")
		urlValues.Add("pass", "vvgo-pass")
		resp, err := noFollow(http.DefaultClient).PostForm(ts.URL, urlValues)
		require.NoError(t, err, "http.PostForm")
		assert.Equal(t, http.StatusFound, resp.StatusCode)
	})

	t.Run("post/success+repeat", func(t *testing.T) {
		loginHandler.Sessions.Init(context.Background())
		ts := httptest.NewServer(loginHandler)
		defer ts.Close()

		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		require.NoError(t, err, "cookiejar.New")
		client := noFollow(&http.Client{Jar: jar})

		urlValues := make(url.Values)
		urlValues.Add("user", "vvgo-user")
		urlValues.Add("pass", "vvgo-pass")
		resp, err := client.PostForm(ts.URL, urlValues)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, http.StatusFound, resp.StatusCode)

		resp, err = client.Get(ts.URL)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/", resp.Header.Get("Location"), "location")
	})
}

func noFollow(client *http.Client) *http.Client {
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return client
}
