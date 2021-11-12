package http

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/clients/http_util"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	server := NewServer("0.0.0.0:8080")
	ts := httptest.NewServer(http.HandlerFunc(server.Server.Handler.ServeHTTP))
	defer ts.Close()

	newRequest := func(t *testing.T, method, url string, roles ...auth.Role) *http.Request {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err, "http.NewRequest")
		if len(roles) != 0 {
			session, err := auth.NewSession(context.Background(), &auth.Identity{
				Roles: roles,
			}, 3600*time.Second)
			require.NoError(t, err, "login.NewSession")
			req.Header.Set("Authorization", "Bearer "+session)
		}
		return req
	}

	doRequest := func(t *testing.T, req *http.Request) *http.Response {
		resp, err := http_util.NoFollow(http.DefaultClient).Do(req)
		require.NoError(t, err, "http.Do")
		return resp
	}

	t.Run("download", func(t *testing.T) {
		t.Run("anonymous", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/download")
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
		t.Run("vvgo-member", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/download", auth.RoleDownload)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})

	t.Run("authorize", func(t *testing.T) {
		t.Run("ok /authorize/vvgo-leader", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-leader", auth.RoleVVGOExecutiveDirector)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("fail /authorize/vvgo-leader", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-leader", auth.RoleVVGOProductionTeam)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
		t.Run("ok /authorize/vvgo-teams", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-teams", auth.RoleVVGOProductionTeam)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("fail /authorize/vvgo-teams", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-teams", auth.RoleVVGOVerifiedMember)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
		t.Run("ok /authorize/vvgo-member", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-member", auth.RoleVVGOVerifiedMember)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
		t.Run("fail /authorize/vvgo-member", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/authorize/vvgo-member", auth.RoleAnonymous)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	})
}
