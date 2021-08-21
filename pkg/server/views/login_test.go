package views

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/models"
	login2 "github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
	"net/http/httptest"
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
			loginView.ServeHTTP(w, r.Clone(context.WithValue(ctx, login2.CtxKeyVVGOIdentity, &models.Identity{Roles: []models.Role{models.RoleVVGOMember}})))
		}))
		defer ts.Close()

		cookie, err := login2.NewCookie(ctx, &models.Identity{Roles: []models.Role{models.RoleVVGOMember}}, 600*time.Second)
		require.NoError(t, err, "sessions.NewCookie()")

		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		require.NoError(t, err, "http.NewRequest()")
		req.AddCookie(cookie)
		resp, err := http_wrappers.NoFollow(nil).Do(req)
		require.NoError(t, err, "http.Do()")
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/login/success", resp.Header.Get("Location"))
	})
}
