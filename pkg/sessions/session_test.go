package sessions

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestSession_String(t *testing.T) {
	secret := Secret{0x560febda7eae12b8, 0xc0cecc7851ca8906, 0x2623d26de389ebcb, 0x5a3097fc6ef622a1}
	session := Session{
		ID:      0x7b7cc95133c4265d,
		Expires: time.Unix(0, 0x1607717a7c5d32e1),
	}
	got := session.String(secret)
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
	require.NoError(t, gotSession.ReadCookie(secret, &src), "Read()")
	assert.Equal(t, wantSession, gotSession, "session")
}

func TestSecret(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		token := NewSecret()
		assert.NoError(t, token.Validate(), "validate")
		assert.NotEqual(t, token.String(), NewSecret().String())
	})
	t.Run("decode", func(t *testing.T) {
		arg := "196ddf804c7666d48d32ff4a91a530bcc5c7cde4a26096ad67758135226bfb2e"
		expected := Secret{0x196ddf804c7666d4, 0x8d32ff4a91a530bc, 0xc5c7cde4a26096ad, 0x67758135226bfb2e}
		var got Secret
		assert.NoError(t, got.Decode(arg))
		assert.Equal(t, expected, got)
	})
	t.Run("string", func(t *testing.T) {
		expected := "196ddf804c7666d48d32ff4a91a530bcc5c7cde4a26096ad67758135226bfb2e"
		arg := Secret{0x196ddf804c7666d4, 0x8d32ff4a91a530bc, 0xc5c7cde4a26096ad, 0x67758135226bfb2e}
		got := arg.String()
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
