package login

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

const CookieName = "vvgo-sessions"
const CookiePath = "/"

func CookieDomain() string {
	x, _ := url.Parse(parse_config.Config.VVGO.ServerUrl)
	return "." + x.Hostname()
}

const CtxKeyVVGOIdentity = "vvgo_identity"

func IdentityFromContext(ctx context.Context) *models.Identity {
	ctxIdentity := ctx.Value(CtxKeyVVGOIdentity)
	identity, ok := ctxIdentity.(*models.Identity)
	if !ok {
		identity = new(models.Identity)
		*identity = models.Anonymous()
	}
	return identity
}

// ReadSessionFromRequest reads the identity from the sessions db based on the request data.
func ReadSessionFromRequest(ctx context.Context, r *http.Request, dest *models.Identity) error {
	bearer := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(bearer, "Bearer ") {
		return GetSession(ctx, bearer[len("Bearer "):], dest)
	}

	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return err
	}
	return GetSession(ctx, cookie.Value, dest)
}

func DeleteSessionFromRequest(ctx context.Context, r *http.Request) error {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return nil
	}
	return DeleteSession(ctx, cookie.Value)
}

// NewCookie returns cookie with a crypto-rand session id.
func NewCookie(ctx context.Context, src *models.Identity, expires time.Duration) (*http.Cookie, error) {
	session, err := NewSession(ctx, src, expires)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     CookieName,
		Value:    session,
		Expires:  time.Now().Add(expires),
		Domain:   CookieDomain(),
		Path:     CookiePath,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
	}, nil
}

func NewCookieValue() string {
	buf := make([]byte, 8)
	result := "V-i-r-t-u-a-l--V-G-O--"
	for i := 0; i < 4; i++ {
		_, _ = rand.Reader.Read(buf)
		result += fmt.Sprintf("%013s", strconv.FormatUint(binary.BigEndian.Uint64(buf), 36))
	}
	return result
}

// NewSession returns a new session with a crypto-rand session id.
func NewSession(ctx context.Context, identity *models.Identity, expires time.Duration) (string, error) {
	value := NewCookieValue()
	key := "sessions:" + value
	stringExpires := strconv.Itoa(int(expires.Seconds()))
	srcBytes, _ := json.Marshal(identity)
	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", key, stringExpires, string(srcBytes))); err != nil {
		return "", err
	}
	return value, nil
}

// GetSession reads the login identity for the given session ID.
func GetSession(ctx context.Context, id string, dest *models.Identity) error {
	var gotBytes []byte
	err := redis.Do(ctx, redis.Cmd(&gotBytes, "GET", "sessions:"+id))
	switch {
	case err != nil:
		return err
	case len(gotBytes) == 0:
		return ErrSessionNotFound
	default:
		return json.NewDecoder(bytes.NewReader(gotBytes)).Decode(dest)
	}
}

// DeleteSession deletes the sessionID key from redis.
func DeleteSession(ctx context.Context, id string) error {
	return redis.Do(ctx, redis.Cmd(nil, "DEL", "sessions:"+id))
}
