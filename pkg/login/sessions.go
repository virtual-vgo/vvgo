package login

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
	"strconv"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

var logger = log.Logger()

// Store provides access to the map of session id's to access roles.
// It can read and validate session cookies from incoming requests,
type Store struct {
	config Config
}

type Config struct {
	// CookieName is the name of the cookies created by the store.
	CookieName string `redis:"cookie_name" default:"vvgo-sessions"`

	// CookieDomain is the domain where the cookies can be used.
	// This should be the domain that users visit in their browser.
	CookieDomain string `redis:"cookie_domain" default:"localhost"`

	// CookiePath is the url path where the cookies can be used.
	CookiePath string `redis:"cookie_path" default:"/"`
}

func newConfig(ctx context.Context) Config {
	var dest Config
	parse_config.SetDefaults(&dest)
	if err := parse_config.ReadFromRedisHash(ctx, &dest, "config:login"); err != nil {
		logger.WithError(err).Errorf("redis.Do() failed: %v", err)
	}
	return dest
}

// NewStore returns a new sessions client.
func NewStore(ctx context.Context) *Store { return &Store{config: newConfig(ctx)} }

func (x *Store) Config() Config { return x.config }

// ReadSessionFromRequest reads the identity from the sessions db based on the request data.
func (x *Store) ReadSessionFromRequest(ctx context.Context, r *http.Request, dest *Identity) error {
	cookie, err := r.Cookie(x.config.CookieName)
	if err != nil {
		return err
	}
	return x.GetSession(ctx, cookie.Value, dest)
}

func (x *Store) DeleteSessionFromRequest(ctx context.Context, r *http.Request) error {
	cookie, err := r.Cookie(x.config.CookieName)
	if err != nil {
		return nil
	}
	return x.DeleteSession(ctx, cookie.Value)
}

// NewCookie returns cookie with a crypto-rand session id.
func (x *Store) NewCookie(ctx context.Context, src *Identity, expires time.Duration) (*http.Cookie, error) {
	session, err := x.NewSession(ctx, src, expires)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     x.config.CookieName,
		Value:    session,
		Expires:  time.Now().Add(expires),
		Domain:   x.config.CookieDomain,
		Path:     x.config.CookiePath,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
	}, nil
}

func NewCookieValue() string {
	buf := make([]byte, 8)
	result := "V-i-r-t-u-a-l--V-G-O--"
	for i := 0; i < 4; i++ {
		rand.Reader.Read(buf)
		result += fmt.Sprintf("%013s", strconv.FormatUint(binary.BigEndian.Uint64(buf), 36))
	}
	return result
}

// NewSession returns a new session with a crypto-rand session id.
func (x *Store) NewSession(ctx context.Context, src *Identity, expires time.Duration) (string, error) {
	value := NewCookieValue()
	key := "sessions:" + value
	stringExpires := strconv.Itoa(int(expires.Seconds()))
	srcBytes, _ := json.Marshal(src)
	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", key, stringExpires, string(srcBytes))); err != nil {
		return "", err
	}
	return value, nil
}

// GetSession reads the login identity for the given session ID.
func (x *Store) GetSession(ctx context.Context, id string, dest *Identity) error {
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
func (x *Store) DeleteSession(ctx context.Context, id string) error {
	return redis.Do(ctx, redis.Cmd(nil, "DEL", "sessions:"+id))
}
