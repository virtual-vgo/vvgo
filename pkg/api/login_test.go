package api

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestLoginHandler_ServeHTTP(t *testing.T) {
	t.Run("failure", func(t *testing.T) {
		ts := httptest.NewServer(new(LoginHandler))
		defer ts.Close()

		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		require.NoError(t, err, "cookiejar.New")
		client := &http.Client{
			Jar: jar,
		}

		resp, err := client.Get(ts.URL)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, resp.StatusCode, http.StatusUnauthorized)
		var gotBody bytes.Buffer
		gotBody.ReadFrom(resp.Body)
		assert.Equal(t, "authorization failed", strings.TrimSpace(gotBody.String()), "body")
	})

	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(new(LoginHandler))
		defer ts.Close()
		tsRealURL, err := url.Parse(ts.URL)
		require.NoError(t, err, "url.Parse")

		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		require.NoError(t, err, "cookiejar.New")
		client := &http.Client{
			Jar: jar,
		}
		t.Log("current cookies:")
		for _, cookie := range jar.Cookies(tsRealURL) {
			t.Logf("%s: %s", cookie.Name, cookie.Value)
		}

		urlValues := make(url.Values)
		urlValues.Add("user", "jackson")
		urlValues.Add("pass", "jackson")
		resp, err := client.PostForm(ts.URL, urlValues)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var gotBody bytes.Buffer
		gotBody.ReadFrom(resp.Body)
		assert.Equal(t, "welcome jackson, have a cookie!", strings.TrimSpace(gotBody.String()), "body")

		var cookieString string
		for _, cookie := range jar.Cookies(tsRealURL) {
			cookieString += fmt.Sprintf("%s: %s\n", cookie.Name, cookie.Value)
		}
		t.Logf("current cookies: %s", cookieString)
	})

	t.Run("success+repeat", func(t *testing.T) {
		ts := httptest.NewServer(new(LoginHandler))
		defer ts.Close()

		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		require.NoError(t, err, "cookiejar.New")
		client := &http.Client{
			Jar: jar,
		}

		urlValues := make(url.Values)
		urlValues.Add("user", "jackson")
		urlValues.Add("pass", "jackson")
		resp, err := client.PostForm(ts.URL, urlValues)
		require.NoError(t, err, "client.PostForm")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var gotBody bytes.Buffer
		gotBody.ReadFrom(resp.Body)
		assert.Equal(t, "welcome jackson, have a cookie!", strings.TrimSpace(gotBody.String()), "body")

		resp, err = client.Get(ts.URL)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		gotBody.Reset()
		gotBody.ReadFrom(resp.Body)
		assert.Equal(t, "welcome back jackson", strings.TrimSpace(gotBody.String()), "body")
	})
}
