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

var logger = log.New()

const ConfigModule = "login"

type Config struct {
	// CookieName is the name of the cookies created by the store.
	CookieName string `redis:"cookie_name" default:"vvgo-sessions"`

	// CookieDomain is the domain where the cookies can be used.
	// This should be the domain that users visit in their browser.
	CookieDomain string `redis:"cookie_domain" default:""`

	// CookiePath is the url path where the cookies can be used.
	CookiePath string `redis:"cookie_path" default:"/"`
}

func readConfig(ctx context.Context) Config {
	var config Config
	parse_config.ReadConfigModule(ctx, ConfigModule, &config)
	parse_config.SetDefaults(&config)
	return config
}

func CookieDomain(ctx context.Context) string {
	return readConfig(ctx).CookieDomain
}

// ReadSessionFromRequest reads the identity from the sessions db based on the request data.
func ReadSessionFromRequest(ctx context.Context, r *http.Request, dest *Identity) error {
	config := readConfig(ctx)
	cookie, err := r.Cookie(config.CookieName)
	if err != nil {
		return err
	}
	return GetSession(ctx, cookie.Value, dest)
}

func DeleteSessionFromRequest(ctx context.Context, r *http.Request) error {
	config := readConfig(ctx)
	cookie, err := r.Cookie(config.CookieName)
	if err != nil {
		return nil
	}
	return DeleteSession(ctx, cookie.Value)
}

// NewCookie returns cookie with a crypto-rand session id.
func NewCookie(ctx context.Context, src *Identity, expires time.Duration) (*http.Cookie, error) {
	config := readConfig(ctx)
	session, err := NewSession(ctx, src, expires)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     config.CookieName,
		Value:    session,
		Expires:  time.Now().Add(expires),
		Domain:   config.CookieDomain,
		Path:     config.CookiePath,
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
func NewSession(ctx context.Context, src *Identity, expires time.Duration) (string, error) {
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
func GetSession(ctx context.Context, id string, dest *Identity) error {
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
