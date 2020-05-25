package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestNewServerAuthorization(t *testing.T) {
	server := NewServer(context.Background(), ServerConfig{
		MemberUser:        "vvgo-member",
		MemberPass:        "vvgo-member",
		UploaderToken:     "vvgo-uploader",
		DeveloperToken:    "vvgo-developer",
		DistroBucketName:  "vvgo-distro" + strconv.Itoa(lrand.Int()),
		BackupsBucketName: "vvgo-backups" + strconv.Itoa(lrand.Int()),
		RedisNamespace:    "vvgo-testing" + strconv.Itoa(lrand.Int()),
		Login: login.Config{
			CookieName: "vvgo-cookie",
		},
	})
	ts := httptest.NewServer(http.HandlerFunc(server.Server.Handler.ServeHTTP))
	defer ts.Close()

	newRequest := func(t *testing.T, method, url string, role login.Role) *http.Request {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err, "http.NewRequest")
		if role != login.RoleAnonymous {
			cookie, err := server.database.Sessions.NewCookie(context.Background(), &login.Identity{
				Roles: []login.Role{role},
			}, 3600*time.Second)
			require.NoError(t, err, "sessions.NewCookie")
			req.AddCookie(cookie)
		}
		return req
	}

	doRequest := func(t *testing.T, req *http.Request) *http.Response {
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "http.Do")
		return resp
	}

	t.Run("index", func(t *testing.T) {
		req := newRequest(t, http.MethodGet, ts.URL, login.RoleAnonymous)
		resp := doRequest(t, req)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("parts", func(t *testing.T) {
		t.Run("anonymous", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/parts", login.RoleAnonymous)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
		t.Run("vvgo-member", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/parts", login.RoleVVGOMember)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})

	t.Run("download", func(t *testing.T) {
		t.Run("anonymous", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/download", login.RoleAnonymous)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
		t.Run("vvgo-member", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/download", login.RoleVVGOMember)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	})

	t.Run("projects", func(t *testing.T) {
		t.Run("anonymous", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/projects", login.RoleAnonymous)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
		t.Run("vvgo-member", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/projects", login.RoleVVGOMember)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
		t.Run("vvgo-uploader", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/projects", login.RoleVVGOUploader)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})

	t.Run("backups", func(t *testing.T) {
		t.Run("anonymous", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/backups", login.RoleAnonymous)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
		t.Run("vvgo-member", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/backups", login.RoleVVGOMember)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
		t.Run("vvgo-uploader", func(t *testing.T) {
			req := newRequest(t, http.MethodGet, ts.URL+"/backups", login.RoleVVGOUploader)
			resp := doRequest(t, req)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})
}
