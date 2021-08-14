package login

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() {
	redis.InitializeFromEnv()
}

func TestStore_GetIdentity(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		ctx := context.Background()
		session, err := NewSession(ctx, &Identity{Kind: "Testing", Roles: []Role{"Tester"}}, 30*time.Second)
		require.NoError(t, err)
		var gotIdentity Identity
		require.NoError(t, GetSession(ctx, session, &gotIdentity))
		assert.Equal(t, Identity{Kind: "Testing", Roles: []Role{"Tester"}}, gotIdentity)
	})

	t.Run("doesnt exist", func(t *testing.T) {
		ctx := context.Background()
		var gotIdentity Identity
		assert.Equal(t, ErrSessionNotFound, GetSession(ctx, "cheese", &gotIdentity))
	})
}

func TestStore_DeleteIdentity(t *testing.T) {
	ctx := context.Background()

	session1, err := NewSession(ctx, &Identity{Kind: "Testing", Roles: []Role{"Tester"}}, 30*time.Second)
	require.NoError(t, err)
	require.NoError(t, DeleteSession(ctx, session1))
	var gotIdentity Identity
	assert.Equal(t, ErrSessionNotFound, GetSession(ctx, session1, &gotIdentity))
}

func TestStore_ReadSessionFromRequest(t *testing.T) {
	t.Run("no session", func(t *testing.T) {
		ctx := context.Background()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  CookieName,
			Value: "cheese",
		})
		var got Identity
		require.Equal(t, ErrSessionNotFound, ReadSessionFromRequest(ctx, req, &got))
	})
	t.Run("cookie", func(t *testing.T) {
		ctx := context.Background()
		session, err := NewSession(ctx, &Identity{Kind: "Testing", Roles: []Role{"Tester"}}, 30*time.Second)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  CookieName,
			Value: session,
		})
		var got Identity
		require.NoError(t, ReadSessionFromRequest(ctx, req, &got))
		assert.Equal(t, Identity{Kind: "Testing", Roles: []Role{"Tester"}}, got)
	})
}

func TestStore_DeleteSessionFromRequest(t *testing.T) {
	t.Run("no session", func(t *testing.T) {
		ctx := context.Background()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, DeleteSessionFromRequest(ctx, req))
	})
	t.Run("cookie", func(t *testing.T) {
		ctx := context.Background()
		session, err := NewSession(ctx, &Identity{Kind: "Testing", Roles: []Role{"Tester"}}, 30*time.Second)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  CookieName,
			Value: session,
		})
		require.NoError(t, DeleteSessionFromRequest(ctx, req))
		var gotIdentity Identity
		assert.Equal(t, ErrSessionNotFound, GetSession(ctx, session, &gotIdentity))
	})
}

func TestStore_NewCookie(t *testing.T) {
	ctx := context.Background()
	gotCookie, err := NewCookie(ctx, &Identity{Kind: "Testing", Roles: []Role{"Tester"}}, 30*time.Second)
	require.NoError(t, err)

	assert.Equal(t, CookieName, gotCookie.Name, "cookie.Name")
	assert.NotEmpty(t, gotCookie.Value, "cookie.Value")
	assert.Equal(t, CookiePath, gotCookie.Path, "cookie.Path")
	assert.Equal(t, ".vvgo.org", gotCookie.Domain, "cookie.Domain")
	assert.Equal(t, true, gotCookie.HttpOnly, "cookie.HttpOnly")
	assert.Equal(t, http.SameSiteStrictMode, gotCookie.SameSite, "cookie.SameSite")
}
