package login

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
	"strconv"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

// Store provides access to the map of session id's to access roles.
// It can read and validate signed session cookies from incoming requests,
// and create new signed cookies for authenticated users.
type Store struct {
	config Config
}

type Config struct {
	// CookieName is the name of the cookies created by the store.
	CookieName string `split_words:"true" default:"vvgo-sessions"`

	// CookieDomain is the domain where the cookies can be used.
	// This should be the domain that users visit in their browser.
	CookieDomain string `split_words:"true" default:"localhost"`

	// CookiePath is the url path where the cookies can be used.
	CookiePath string `split_words:"true" default:"/"`

	// Namespace is prefixed to all redis keys.
	Namespace string `split_words:"true"`
}

const DataFile = "users.json"

// NewStore returns a new sessions client.
func NewStore(config Config) *Store {
	return &Store{
		config: config,
	}
}

// ReadSessionFromRequest reads the identity from the sessions db based on the request data.
func (x *Store) ReadSessionFromRequest(ctx context.Context, r *http.Request, dest *Identity) error {
	cookie, err := r.Cookie(x.config.CookieName)
	if err != nil {
		return err
	}
	return x.GetSession(ctx, cookie.Value, dest)
}

// NewCookie returns cookie with a cryptographically signed session payload.
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

// NewSession returns a new session with a crypto-rand session id.
func (x *Store) NewSession(ctx context.Context, src *Identity, expires time.Duration) (string, error) {
	buf := make([]byte, 8)
	result := "V-i-r-t-u-a-l--V-G-O--"
	for i := 0; i < 4; i++ {
		rand.Reader.Read(buf)
		result += fmt.Sprintf("%013s", strconv.FormatUint(binary.BigEndian.Uint64(buf), 36))
	}

	key := x.config.Namespace + ":sessions:" + result
	if err := redis.Do(ctx, redis.FlatCmd(nil, "HSET", key, src)); err != nil {
		return "", err
	}
	if err := redis.Do(ctx, redis.Cmd(nil, "EXPIRE", key, strconv.Itoa(int(expires.Seconds())))); err != nil {
		return "", err
	}
	return result, nil
}

// GetSession reads the login identity for the given session ID.
func (x *Store) GetSession(ctx context.Context, id string, dest *Identity) error {
	err := redis.Do(ctx, redis.Cmd(dest, "HGETALL", x.config.Namespace+":sessions:"+id))
	switch {
	case err != nil:
		return err
	case dest.Kind == "":
		return ErrSessionNotFound
	default:
		return nil
	}
}

// DeleteSession deletes the sessionID key from the map.
func (x *Store) DeleteSession(ctx context.Context, id string) error {
	return redis.Do(ctx, redis.Cmd(nil, "DEL", x.config.Namespace+":sessions:"+id))
}
