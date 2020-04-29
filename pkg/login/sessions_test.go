package login

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net/http"
	"net/http/httptest"
	"strings"
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
		// storage is uninitialized, so this should throw an error.
		var gotIdentity Identity
		require.Error(t, store.GetIdentity(ctx, 42069, &gotIdentity))
	})
}

func TestStore_StoreIdentity(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		store := NewStore(locker.NewLocksmith(locker.Config{}), Config{})
		require.NoError(t, store.Init(ctx))

		// add an existing session
		require.NoError(t, store.cache.PutObject(ctx, DataFile,
			storage.NewJSONObject([]byte(`{"2020":{"kind":"Rapture","roles":["EndTimes"]}}`))))

		// store the new session
		require.NoError(t, store.StoreIdentity(ctx, 42069, &Identity{
			Kind:  "Test",
			Roles: []Role{"Tester"},
		}))

		var gotObject storage.Object
		store.cache.GetObject(ctx, DataFile, &gotObject)
		assert.Equal(t, "application/json", gotObject.ContentType)
		assert.Equal(t, `{"2020":{"kind":"Rapture","roles":["EndTimes"]},"42069":{"kind":"Test","roles":["Tester"]}}`,
			strings.TrimSpace(string(gotObject.Bytes)))
	})

	t.Run("storage error", func(t *testing.T) {
		ctx := context.Background()
		store := NewStore(locker.NewLocksmith(locker.Config{}), Config{})
		// storage is uninitialized, so this should throw an error.
		var gotIdentity Identity
		require.Error(t, store.StoreIdentity(ctx, 42069, &gotIdentity))
	})
}

func TestStore_DeleteIdentity(t *testing.T) {
	ctx := context.Background()
	store := NewStore(locker.NewLocksmith(locker.Config{}), Config{})
	require.NoError(t, store.Init(ctx))
	require.NoError(t, store.cache.PutObject(ctx, DataFile,
		storage.NewJSONObject([]byte(`{
			"2020":{"kind":"Rapture","roles": ["EndTimes"]},
			"42069":{"kind":"Test","roles": ["Tester"]}
		}`))))

	require.NoError(t, store.DeleteIdentity(ctx, 42069))

	var gotObject storage.Object
	store.cache.GetObject(ctx, DataFile, &gotObject)
	assert.Equal(t, "application/json", gotObject.ContentType)
	assert.Equal(t, `{"2020":{"kind":"Rapture","roles":["EndTimes"]}}`,
		strings.TrimSpace(string(gotObject.Bytes)))
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

func TestStore_NewSession(t *testing.T) {
	t.Run("are valid", func(t *testing.T) {
		store := NewStore(locker.NewLocksmith(locker.Config{}), Config{})
		session := store.NewSession(time.Unix(0, 0x2642f3cd1200))
		assert.Equal(t, uint64(0x2642f3cd1200), session.ExpiresNanos)
		assert.NotEqual(t, 0, session.ID)
	})

	t.Run("unique ids", func(t *testing.T) {
		store := NewStore(locker.NewLocksmith(locker.Config{}), Config{})
		for i := 0; i < 100; i++ {
			session := store.NewSession(time.Unix(42069, 0))
			assert.NotEqual(t, store.NewSession(time.Unix(42069, 0)), session)
		}
	})
}

func TestStore_NewCookie(t *testing.T) {
	secret := Secret{0x560febda7eae12b8, 0xc0cecc7851ca8906, 0x2623d26de389ebcb, 0x5a3097fc6ef622a1}
	store := NewStore(locker.NewLocksmith(locker.Config{}), Config{
		Secret:       secret,
		CookiePath:   "/authorized",
		CookieName:   "cookie-name",
		CookieDomain: "tester.local",
	})
	session := Session{
		ID:           0x7b7cc95133c4265d,
		ExpiresNanos: uint64(4743644400 * time.Second),
	}
	gotCookie := store.NewCookie(session)
	wantCookie := &http.Cookie{
		Name:     "cookie-name",
		Value:    "V-i-r-t-u-a-l--V-G-O--16b29700a96cd2cf48b91041e552f3f4b3ce87f1c75cb621ca7f97619ce2f88d7b7cc95133c4265d41d4cf72eac56000",
		Path:     "/authorized",
		Domain:   "tester.local",
		Expires:  time.Unix(4743644400, 0),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	assert.Equal(t, wantCookie, gotCookie)
}

func TestSession_Encode(t *testing.T) {
	secret := Secret{0x560febda7eae12b8, 0xc0cecc7851ca8906, 0x2623d26de389ebcb, 0x5a3097fc6ef622a1}
	session := Session{
		ID:           0x7b7cc95133c4265d,
		ExpiresNanos: uint64(4743644400 * time.Second),
	}
	got := session.Encode(secret)
	want := "V-i-r-t-u-a-l--V-G-O--16b29700a96cd2cf48b91041e552f3f4b3ce87f1c75cb621ca7f97619ce2f88d7b7cc95133c4265d41d4cf72eac56000"
	assert.Equal(t, want, got, "value")
}

func TestSession_Decode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		secret := Secret{0x560febda7eae12b8, 0xc0cecc7851ca8906, 0x2623d26de389ebcb, 0x5a3097fc6ef622a1}
		src := "V-i-r-t-u-a-l--V-G-O--16b29700a96cd2cf48b91041e552f3f4b3ce87f1c75cb621ca7f97619ce2f88d7b7cc95133c4265d41d4cf72eac56000"
		wantSession := Session{
			ID:           0x7b7cc95133c4265d,
			ExpiresNanos: uint64(4743644400 * time.Second),
		}
		t.Log("want session:", wantSession.Encode(secret))

		var gotSession Session
		assert.Equal(t, nil, gotSession.Decode(secret, src), "Read()")
		assert.Equal(t, wantSession.ID, gotSession.ID, "session.ID")
		assert.Equal(t, wantSession.ExpiresNanos, gotSession.ExpiresNanos, "session.ExpiresNanos")
	})

	t.Run("invalid session", func(t *testing.T) {
		secret := Secret{0x560febda7eae12b8, 0xc0cecc7851ca8906, 0x2623d26de389ebcb, 0x5a3097fc6ef622a1}
		src := "hamster wheel"
		wantSession := Session{}
		var gotSession Session
		assert.Equal(t, ErrInvalidSession, gotSession.Decode(secret, src), "Read()")
		assert.Equal(t, wantSession, gotSession, "session")
	})

	t.Run("invalid signature and expired", func(t *testing.T) {
		secret := Secret{0x560febda7eae12b8, 0xc0cecc7851ca8906, 0x2623d26de389ebcb, 0x5a3097fc6ef622a1}
		src := "V-i-r-t-u-a-l--V-G-O--00000000b67f71856e83d9aa5ef4644d76c2a8736c440c6851861172e44ae7b07b7cc95133c4265d1607717a7c5d32e1"
		wantSession := Session{
			ID:           0x7b7cc95133c4265d,
			ExpiresNanos: 0x1607717a7c5d32e1,
		}
		t.Log("want session:", wantSession.Encode(secret))

		var gotSession Session
		assert.Equal(t, ErrInvalidSignature, gotSession.Decode(secret, src), "Read()")
		assert.Equal(t, wantSession, gotSession, "session")
	})

	t.Run("expired", func(t *testing.T) {
		secret := Secret{0x560febda7eae12b8, 0xc0cecc7851ca8906, 0x2623d26de389ebcb, 0x5a3097fc6ef622a1}
		src := "V-i-r-t-u-a-l--V-G-O--7f13099d9edb1fa2e54d58404e63ba1a9dd1afbdbeea3f83576cc40232fc9bae7b7cc95133c4265d1607717a7c5d32e1"
		wantSession := Session{
			ID:           0x7b7cc95133c4265d,
			ExpiresNanos: 0x1607717a7c5d32e1,
		}

		var gotSession Session
		assert.Equal(t, ErrSessionExpired, gotSession.Decode(secret, src), "Read()")
		assert.Equal(t, wantSession, gotSession, "session")
	})
}

func TestSession_DecodeCookie(t *testing.T) {
	secret := Secret{0x560febda7eae12b8, 0xc0cecc7851ca8906, 0x2623d26de389ebcb, 0x5a3097fc6ef622a1}
	src := http.Cookie{
		Value: "V-i-r-t-u-a-l--V-G-O--16b29700a96cd2cf48b91041e552f3f4b3ce87f1c75cb621ca7f97619ce2f88d7b7cc95133c4265d41d4cf72eac56000",
	}
	wantSession := Session{
		ID:           0x7b7cc95133c4265d,
		ExpiresNanos: 0x41d4cf72eac56000,
	}

	var gotSession Session
	assert.Equal(t, nil, gotSession.DecodeCookie(secret, &src), "Read()")
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
