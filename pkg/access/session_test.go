package access

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStore_Init(t *testing.T) {
	ctx := context.Background()
	store := NewStore(locker.NewLocksmith(locker.Config{}), Config{})
	require.NoError(t, store.Init(ctx))
	var gotObj storage.Object
	store.cache.GetObject(ctx, DataFile, &gotObj)
	assert.Equal(t, storage.Object{
		ContentType: "application/json",
		Bytes:       []byte(`{}`),
	}, gotObj)
}

func TestStore_GetIdentity(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		ctx := context.Background()
		store := NewStore(locker.NewLocksmith(locker.Config{}), Config{})
		require.NoError(t, store.Init(ctx))
		require.NoError(t, store.cache.PutObject(ctx, DataFile,
			storage.NewJSONObject([]byte(`{"42069":{"kind":"Testing","roles":["Tester"]}}`))))

		var gotIdentity Identity
		require.NoError(t, store.GetIdentity(ctx, 42069, &gotIdentity))
		assert.Equal(t, Identity{Kind: "Testing", Roles: []Role{"Tester"}}, gotIdentity)
	})

	t.Run("doesnt exist", func(t *testing.T) {
		ctx := context.Background()
		store := NewStore(locker.NewLocksmith(locker.Config{}), Config{})
		require.NoError(t, store.Init(ctx))

		var gotIdentity Identity
		assert.Equal(t, ErrSessionNotFound, store.GetIdentity(ctx, 42069, &gotIdentity))
	})

	t.Run("storage error", func(t *testing.T) {
		ctx := context.Background()
		store := NewStore(locker.NewLocksmith(locker.Config{}), Config{})

		var gotIdentity Identity
		require.Error(t, store.GetIdentity(ctx, 42069, &gotIdentity))
	})
}

func TestStore_DeleteIdentity(t *testing.T) {

}

func TestStore_StoreIdentity(t *testing.T) {

}

func TestStore_ReadSessionFromRequest(t *testing.T) {
	secret := Secret{0x560febda7eae12b8, 0xc0cecc7851ca8906, 0x2623d26de389ebcb, 0x5a3097fc6ef622a1}
	store := NewStore(locker.NewLocksmith(locker.Config{}), Config{Secret: secret, CookieName: "vvgo-cookie"})
	session := store.NewSession(time.Now().Add(1e6 * time.Second))
	t.Run("no session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		var gotSession Session
		require.Equal(t, ErrSessionNotFound, store.ReadSessionFromRequest(req, &gotSession))
	})
	t.Run("bearer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", "Bearer "+session.Encode(secret))
		var gotSession Session
		require.NoError(t, store.ReadSessionFromRequest(req, &gotSession))
		assert.Equal(t, session.ID, gotSession.ID)
	})
	t.Run("cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  "vvgo-cookie",
			Value: session.Encode(secret),
		})
		var gotSession Session
		require.NoError(t, store.ReadSessionFromRequest(req, &gotSession))
		assert.Equal(t, session.ID, gotSession.ID)
	})
}

func TestSession_String(t *testing.T) {
	secret := Secret{0x560febda7eae12b8, 0xc0cecc7851ca8906, 0x2623d26de389ebcb, 0x5a3097fc6ef622a1}
	session := Session{
		ID:      0x7b7cc95133c4265d,
		Expires: time.Unix(0, 0x1607717a7c5d32e1),
	}
	got := session.Encode(secret)
	wantCookieValue := "V-i-r-t-u-a-l--V-G-O--01677fc8b67f71856e83d9aa5ef4644d76c2a8736c440c6851861172e44ae7b07b7cc95133c4265d1607717a7c5d32e1"
	assert.Equal(t, wantCookieValue, got, "value")
}

func TestSession_ReadCookie(t *testing.T) {
	secret := Secret{0x560febda7eae12b8, 0xc0cecc7851ca8906, 0x2623d26de389ebcb, 0x5a3097fc6ef622a1}
	src := http.Cookie{
		Value: "V-i-r-t-u-a-l--V-G-O--01677fc8b67f71856e83d9aa5ef4644d76c2a8736c440c6851861172e44ae7b07b7cc95133c4265d1607717a7c5d32e1",
	}
	wantSession := Session{
		ID:      0x7b7cc95133c4265d,
		Expires: time.Unix(0, 0x1607717a7c5d32e1),
	}

	var gotSession Session
	assert.Equal(t, ErrSessionExpired, gotSession.DecodeCookie(secret, &src), "Read()")
	assert.Equal(t, wantSession, gotSession, "session")
}

func TestSecret(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		token := NewSecret()
		assert.NoError(t, token.Validate(), "validate")
		assert.NotEqual(t, token.String(), NewSecret().String())
	})
	t.Run("string", func(t *testing.T) {
		expected := "196ddf804c7666d48d32ff4a91a530bcc5c7cde4a26096ad67758135226bfb2e"
		arg := Secret{0x196ddf804c7666d4, 0x8d32ff4a91a530bc, 0xc5c7cde4a26096ad, 0x67758135226bfb2e}
		got := arg.String()
		assert.Equal(t, expected, got)
	})
	t.Run("decode", func(t *testing.T) {
		arg := "196ddf804c7666d48d32ff4a91a530bcc5c7cde4a26096ad67758135226bfb2e"
		expected := Secret{0x196ddf804c7666d4, 0x8d32ff4a91a530bc, 0xc5c7cde4a26096ad, 0x67758135226bfb2e}
		var got Secret
		assert.NoError(t, got.Decode(arg))
		assert.Equal(t, expected, got)
	})
	t.Run("validate/success", func(t *testing.T) {
		arg := Secret{0x196ddf804c7666d4, 0x8d32ff4a91a530bc, 0xc5c7cde4a26096ad, 0x67758135226bfb2e}
		assert.NoError(t, arg.Validate())
	})
	t.Run("validate/fail", func(t *testing.T) {
		arg := Secret{0, 0x8d32ff4a91a530bc, 0xc5c7cde4a26096ad, 0x67758135226bfb2e}
		assert.Equal(t, ErrInvalidSecret, arg.Validate())
	})
}
