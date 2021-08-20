package server

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRBACMux_Handle(t *testing.T) {
	ctx := context.Background()
	mux := RBACMux{
		ServeMux: http.NewServeMux(),
	}

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// do nothing
	}, login.RoleVVGOTeams)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	t.Run("no auth", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		require.NoError(t, err, "http.NewRequest()")
		resp, err := http_wrappers.NoFollow(nil).Do(req)
		require.NoError(t, err, "http.Do()")
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/login?target=%2F", resp.Header.Get("Location"))
	})

	t.Run("basic auth", func(t *testing.T) {
		mux.Basic = map[[2]string][]login.Role{
			{"uploader", "uploader"}: {login.RoleVVGOTeams},
			{"member", "member"}:     {login.RoleVVGOMember},
		}

		newAuthRequest := func(t *testing.T, user, pass string) *http.Request {
			req, err := http.NewRequest(http.MethodGet, ts.URL, strings.NewReader(""))
			require.NoError(t, err, "http.NewRequest")
			req.SetBasicAuth(user, pass)
			return req
		}

		t.Run("success", func(t *testing.T) {
			req := newAuthRequest(t, "uploader", "uploader")
			resp, err := http_wrappers.NoFollow(nil).Do(req)
			require.NoError(t, err, "http.Do()")
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("incorrect user", func(t *testing.T) {
			req := newAuthRequest(t, "", "uploader")
			resp, err := http_wrappers.NoFollow(nil).Do(req)
			require.NoError(t, err, "http.Do()")
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/login?target=%2F", resp.Header.Get("Location"))
		})
		t.Run("incorrect pass", func(t *testing.T) {
			req := newAuthRequest(t, "uploader", "")
			resp, err := http_wrappers.NoFollow(nil).Do(req)
			require.NoError(t, err, "http.Do()")
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/login?target=%2F", resp.Header.Get("Location"))
		})
		t.Run("incorrect role", func(t *testing.T) {
			req := newAuthRequest(t, "member", "member")
			resp, err := http_wrappers.NoFollow(nil).Do(req)
			require.NoError(t, err, "http.Do()")
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})

	t.Run("token auth", func(t *testing.T) {
		mux.Bearer = map[string][]login.Role{
			"uploader": {login.RoleVVGOTeams},
			"member":   {login.RoleVVGOMember},
		}
		newAuthRequest := func(t *testing.T, token string) *http.Request {
			req, err := http.NewRequest(http.MethodGet, ts.URL, strings.NewReader(""))
			require.NoError(t, err, "http.NewRequest")
			req.Header.Set("Authorization", "Bearer "+token)
			return req
		}

		t.Run("success", func(t *testing.T) {
			req := newAuthRequest(t, "uploader")
			resp, err := http_wrappers.NoFollow(nil).Do(req)
			require.NoError(t, err, "http.Do()")
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("incorrect token", func(t *testing.T) {
			req := newAuthRequest(t, "asdfa")
			resp, err := http_wrappers.NoFollow(nil).Do(req)
			require.NoError(t, err, "http.Do()")
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/login?target=%2F", resp.Header.Get("Location"))
		})
		t.Run("incorrect role", func(t *testing.T) {
			req := newAuthRequest(t, "member")
			resp, err := http_wrappers.NoFollow(nil).Do(req)
			require.NoError(t, err, "http.Do()")
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})

	t.Run("login session", func(t *testing.T) {
		newAuthRequest := func(t *testing.T, identity *login.Identity) *http.Request {
			cookie, err := login.NewCookie(ctx, identity, 3600*time.Second)
			require.NoError(t, err, "NewCookie()")
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, ts.URL, strings.NewReader(""))
			require.NoError(t, err, "http.NewRequest")
			req.AddCookie(cookie)
			return req
		}

		t.Run("success", func(t *testing.T) {
			req := newAuthRequest(t, &login.Identity{
				Roles: []login.Role{login.RoleVVGOTeams},
			})
			resp, err := http_wrappers.NoFollow(nil).Do(req)
			require.NoError(t, err, "http.Do()")
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("incorrect role", func(t *testing.T) {
			req := newAuthRequest(t, &login.Identity{
				Roles: []login.Role{login.RoleVVGOMember},
			})
			resp, err := http_wrappers.NoFollow(nil).Do(req)
			require.NoError(t, err, "http.Do()")
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})
}
