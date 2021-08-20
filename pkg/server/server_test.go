package server

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/server/views"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() { views.PublicFiles = "../../public" }

func TestServer(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	server := NewServer("0.0.0.0:8080")
	ts := httptest.NewServer(http.HandlerFunc(server.Server.Handler.ServeHTTP))
	defer ts.Close()

	newRequest := func(t *testing.T, method, url string, roles ...login.Role) *http.Request {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err, "http.NewRequest")
		if len(roles) != 0 {
			cookie, err := login.NewCookie(context.Background(), &login.Identity{
				Roles: roles,
			}, 3600*time.Second)
			require.NoError(t, err, "sessions.NewCookie")
			req.AddCookie(cookie)
		}
		return req
	}

	doRequest := func(t *testing.T, req *http.Request) *http.Response {
		resp, err := http_wrappers.NoFollow(http.DefaultClient).Do(req)
		require.NoError(t, err, "http.Do")
		return resp
	}

	t.Run("parts", func(t *testing.T) {
		t.Run("anonymous", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/parts")
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/login?target=%2Fparts", resp.Header.Get("Location"))
		})
	})

	t.Run("download", func(t *testing.T) {
		t.Run("anonymous", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/download")
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/login?target=%2Fdownload", resp.Header.Get("Location"))
		})
		t.Run("vvgo-member", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/download", login.RoleVVGOMember)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})

	t.Run("authorize", func(t *testing.T) {
		t.Run("ok /authorize/vvgo-leader", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-leader", login.RoleVVGOLeader)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("fail /authorize/vvgo-leader", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-leader", login.RoleVVGOTeams)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
		t.Run("ok /authorize/vvgo-teams", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-teams", login.RoleVVGOTeams)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("fail /authorize/vvgo-teams", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-teams", login.RoleVVGOMember)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
		t.Run("ok /authorize/vvgo-member", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-member", login.RoleVVGOMember)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("fail /authorize/vvgo-member", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-member", login.RoleAnonymous)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})

	t.Run("roles", func(t *testing.T) {
		t.Run("anonymous", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/api/v1/roles")
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			var got []login.Role
			assert.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
			assert.Equal(t, []login.Role{login.RoleAnonymous}, got)
		})
		t.Run("vvgo-uploader", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/api/v1/roles", login.RoleVVGOTeams, login.RoleVVGOMember)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			var got []login.Role
			assert.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
			assert.Equal(t, []login.Role{login.RoleVVGOTeams, login.RoleVVGOMember}, got)
		})
	})

	t.Run("login", func(t *testing.T) {
		t.Run("anonymous", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/login")
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})

	t.Run("logout", func(t *testing.T) {
		t.Run("anonymous", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/logout")
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "/", resp.Header.Get("Location"))
		})
	})
}
