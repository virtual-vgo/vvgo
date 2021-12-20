package rbac

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestRBACMux_Handle(t *testing.T) {
	ctx := context.Background()
	okHandler := func(r *http.Request) models.ApiResponse { return http_helpers.NewOkResponse() }

	newAnonymousRequest := func() *http.Request {
		return httptest.NewRequest(http.MethodGet, "/", nil)
	}

	newBearerRequest := func(t *testing.T, identity *models.Identity) *http.Request {
		t.Helper()
		session, err := login.NewSession(ctx, identity, 3600*time.Second)
		require.NoError(t, err, "login.NewSession()")
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+session)
		return req
	}

	newTokenRequest := func(t *testing.T, identity *models.Identity) *http.Request {
		t.Helper()
		session, err := login.NewSession(ctx, identity, 3600*time.Second)
		require.NoError(t, err, "login.NewSession()")
		params := make(url.Values)
		params.Set("token", session)
		req := httptest.NewRequest(http.MethodGet, "/?"+params.Encode(), nil)
		return req
	}

	assertSuccess := func(t *testing.T, mux Mux, req *http.Request) {
		t.Helper()
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, req)
		resp := recorder.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	assertUnauthorized := func(t *testing.T, mux Mux, req *http.Request) {
		t.Helper()
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, req)
		resp := recorder.Result()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}

	t.Run("no auth", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			mux := NewRBACMux()
			mux.HandleApiFunc("/", okHandler, models.RoleAnonymous)
			assertSuccess(t, mux, newAnonymousRequest())
		})
		t.Run("incorrect role", func(t *testing.T) {
			mux := NewRBACMux()
			mux.HandleApiFunc("/", okHandler, models.RoleVVGOProductionTeam)
			assertUnauthorized(t, mux, newAnonymousRequest())
		})
	})

	t.Run("bearer", func(t *testing.T) {
		mux := NewRBACMux()
		mux.HandleApiFunc("/", okHandler, models.RoleVVGOProductionTeam)
		t.Run("success", func(t *testing.T) {
			assertSuccess(t, mux, newBearerRequest(t, &models.Identity{
				Roles: []models.Role{models.RoleVVGOProductionTeam},
			}))
		})
		t.Run("incorrect role", func(t *testing.T) {
			assertUnauthorized(t, mux, newBearerRequest(t, &models.Identity{
				Roles: []models.Role{models.RoleVVGOVerifiedMember},
			}))
		})
	})

	t.Run("token", func(t *testing.T) {
		mux := NewRBACMux()
		mux.HandleApiFunc("/", okHandler, models.RoleVVGOProductionTeam)
		t.Run("success", func(t *testing.T) {
			assertSuccess(t, mux, newTokenRequest(t, &models.Identity{
				Roles: []models.Role{models.RoleVVGOProductionTeam},
			}))
		})
		t.Run("incorrect role", func(t *testing.T) {
			assertUnauthorized(t, mux, newTokenRequest(t, &models.Identity{
				Roles: []models.Role{models.RoleVVGOVerifiedMember},
			}))
		})
	})
}
