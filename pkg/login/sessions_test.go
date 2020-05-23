package login

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

var lrand = rand.New(rand.NewSource(time.Now().UnixNano()))

func newStore() *Store {
	return NewStore("testing"+strconv.Itoa(lrand.Int()), Config{})
}

func TestStore_GetIdentity(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		ctx := context.Background()
		store := newStore()
		session, err := store.NewSession(ctx, &Identity{Kind: "Testing", Roles: []Role{"Tester"}}, 30*time.Second)
		require.NoError(t, err)
		var gotIdentity Identity
		require.NoError(t, store.GetSession(ctx, session, &gotIdentity))
		assert.Equal(t, Identity{Kind: "Testing", Roles: []Role{"Tester"}}, gotIdentity)
	})

	t.Run("doesnt exist", func(t *testing.T) {
		ctx := context.Background()
		store := newStore()

		var gotIdentity Identity
		assert.Equal(t, ErrSessionNotFound, store.GetSession(ctx, "cheese", &gotIdentity))
	})
}

func TestStore_DeleteIdentity(t *testing.T) {
	ctx := context.Background()
	store := newStore()

	session1, err := store.NewSession(ctx, &Identity{Kind: "Testing", Roles: []Role{"Tester"}}, 30*time.Second)
	require.NoError(t, err)
	require.NoError(t, store.DeleteSession(ctx, session1))
	var gotIdentity Identity
	assert.Equal(t, ErrSessionNotFound, store.GetSession(ctx, session1, &gotIdentity))
}

func TestStore_ReadSessionFromRequest(t *testing.T) {
	t.Run("no session", func(t *testing.T) {
		ctx := context.Background()
		store := newStore()
		store.config.CookieName = "vvgo-cookie"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "vvgo-cookie",
			Value: "cheese",
		})
		var got Identity
		require.Equal(t, ErrSessionNotFound, store.ReadSessionFromRequest(ctx, req, &got))
	})
	t.Run("cookie", func(t *testing.T) {
		ctx := context.Background()
		store := newStore()
		store.config.CookieName = "vvgo-cookie"
		session, err := store.NewSession(ctx, &Identity{Kind: "Testing", Roles: []Role{"Tester"}}, 30*time.Second)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "vvgo-cookie",
			Value: session,
		})
		var got Identity
		require.NoError(t, store.ReadSessionFromRequest(ctx, req, &got))
		assert.Equal(t, Identity{Kind: "Testing", Roles: []Role{"Tester"}}, got)
	})
}

func TestStore_DeleteSessionFromRequest(t *testing.T) {
	t.Run("no session", func(t *testing.T) {
		ctx := context.Background()
		store := newStore()
		store.config.CookieName = "vvgo-cookie"

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		require.NoError(t, store.DeleteSessionFromRequest(ctx, req))
	})
	t.Run("cookie", func(t *testing.T) {
		ctx := context.Background()
		store := newStore()
		store.config.CookieName = "vvgo-cookie"
		session, err := store.NewSession(ctx, &Identity{Kind: "Testing", Roles: []Role{"Tester"}}, 30*time.Second)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "vvgo-cookie",
			Value: session,
		})
		require.NoError(t, store.DeleteSessionFromRequest(ctx, req))
		var gotIdentity Identity
		assert.Equal(t, ErrSessionNotFound, store.GetSession(ctx, session, &gotIdentity))
	})
}

func TestStore_NewCookie(t *testing.T) {
	ctx := context.Background()
	store := newStore()
	store.config.CookiePath = "/authorized"
	store.config.CookieName = "cookie-name"
	store.config.CookieDomain = "tester.local"
	gotCookie, err := store.NewCookie(ctx, &Identity{Kind: "Testing", Roles: []Role{"Tester"}}, 30*time.Second)
	require.NoError(t, err)

	assert.Equal(t, "cookie-name", gotCookie.Name, "cookie.Name")
	assert.NotEmpty(t, gotCookie.Value, "cookie.Value")
	assert.Equal(t, "/authorized", gotCookie.Path, "cookie.Path")
	assert.Equal(t, "tester.local", gotCookie.Domain, "cookie.Domain")
	assert.Equal(t, true, gotCookie.HttpOnly, "cookie.HttpOnly")
	assert.Equal(t, http.SameSiteStrictMode, gotCookie.SameSite, "cookie.SameSite")
}
