package login

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestGetSession(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		ctx := context.Background()
		session, err := NewSession(ctx, &models.Identity{Kind: "Testing", Roles: []models.Role{"Tester"}}, 30*time.Second)
		require.NoError(t, err)
		var gotIdentity models.Identity
		require.NoError(t, GetSession(ctx, session, &gotIdentity))
		assert.Equal(t, models.Identity{
			Key:       session,
			Kind:      "Testing",
			Roles:     []models.Role{"Tester"},
			ExpiresAt: gotIdentity.ExpiresAt, // TODO: implement a better test
			CreatedAt: gotIdentity.CreatedAt, // TODO: implement a better test
		}, gotIdentity)
	})

	t.Run("doesnt exist", func(t *testing.T) {
		ctx := context.Background()
		var gotIdentity models.Identity
		assert.Equal(t, ErrSessionNotFound, GetSession(ctx, "cheese", &gotIdentity))
	})
}

func TestDeleteSession(t *testing.T) {
	ctx := context.Background()

	session1, err := NewSession(ctx, &models.Identity{Kind: "Testing", Roles: []models.Role{"Tester"}}, 30*time.Second)
	require.NoError(t, err)
	require.NoError(t, DeleteSession(ctx, session1))
	var gotIdentity models.Identity
	assert.Equal(t, ErrSessionNotFound, GetSession(ctx, session1, &gotIdentity))
}

func TestReadSessionFromRequest(t *testing.T) {
	t.Run("no session", func(t *testing.T) {
		ctx := context.Background()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		var got models.Identity
		ReadSessionFromRequest(ctx, req, &got)
		require.Equal(t, models.Anonymous(), got)
	})
	t.Run("Bearer", func(t *testing.T) {
		ctx := context.Background()
		session, err := NewSession(ctx, &models.Identity{
			Kind:  "Testing",
			Roles: []models.Role{"Tester"},
		}, 30*time.Second)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+session)

		var got models.Identity
		ReadSessionFromRequest(ctx, req, &got)
		assert.Equal(t, models.Identity{
			Key:       session,
			Kind:      "Testing",
			Roles:     []models.Role{"Tester"},
			ExpiresAt: got.ExpiresAt, // TODO: Implement a better test here.
			CreatedAt: got.CreatedAt, // TODO: Implement a better test here.
		}, got)
	})
	t.Run("Token", func(t *testing.T) {
		ctx := context.Background()
		session, err := NewSession(ctx, &models.Identity{
			Kind:  "Testing",
			Roles: []models.Role{"Tester"},
		}, 30*time.Second)
		require.NoError(t, err)

		params := make(url.Values)
		params.Set("token", session)
		req := httptest.NewRequest(http.MethodGet, "/?"+params.Encode(), nil)

		var got models.Identity
		ReadSessionFromRequest(ctx, req, &got)
		assert.Equal(t, models.Identity{
			Key:       session,
			Kind:      "Testing",
			Roles:     []models.Role{"Tester"},
			ExpiresAt: got.ExpiresAt, // TODO: Implement a better test here.
			CreatedAt: got.CreatedAt, // TODO: Implement a better test here.
		}, got)
	})
}
