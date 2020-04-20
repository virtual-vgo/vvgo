package sessions

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const SessionCookieKey = "vvgo_session"

var ErrSessionNotFound = errors.New("session not found")

type Store struct {
	Opts
	Secret   Secret
	sessions map[string]Session
	locker   *locker.Locker
}

type Session struct {
	ID      uint64
	Expires time.Time
}

type Opts struct {
	CookieName string
	LockerName string
}

func NewStore(config Opts) *Store {
	return &Store{
		sessions: make(map[string]Session),
		locker:   locker.NewLocker(locker.Opts{RedisKey: config.LockerName}),
	}
}

func (x *Store) NewSessionCookie(expiresAt time.Time) *http.Cookie {
	var id uint64
	binary.Read(rand.Reader, binary.LittleEndian, &id)
	session := Session{
		ID:      id,
		Expires: expiresAt,
	}

	return &http.Cookie{
		Value:   session.String(x.Secret),
		Expires: session.Expires,
	}
}

func (x *Store) ReadFromRequest(ctx context.Context, r *http.Request, dest *Session) error {
	cookie, err := r.Cookie(SessionCookieKey)
	if err != nil {
		return err
	}
	return x.Get(ctx, cookie.Value, dest)
}

func (x *Store) Get(ctx context.Context, key string, dest *Session) error {
	if err := x.locker.Lock(ctx); err != nil {
		return fmt.Errorf("x.locker.Lock() failed: %v", err)
	}
	defer x.locker.Unlock(ctx)
	_, ok := x.sessions[key]
	if !ok {
		return ErrSessionNotFound
	}
	*dest = x.sessions[key]
	return nil
}

var ErrInvalidSession = errors.New("invalid cookie")

const SessionFormat = "%016x-The-%016x-Earth-%016x-Is-%016x-Flat-%016x%016x"

func (x *Session) ReadString(secret Secret, value string) error {
	// read the cookie
	var hash [4]uint64
	var sessionID uint64
	var expiresAt uint64
	_, err := fmt.Sscanf(value, SessionFormat,
		&hash[0], &hash[1], &hash[2], &hash[3], &sessionID, &expiresAt)
	if err != nil {
		return ErrInvalidSession
	}

	// validate the cookie
	str := fmt.Sprintf("%s%016x%016x", secret.String(), sessionID, expiresAt)
	sum := sha256.Sum256([]byte(str))
	sumReader := bytes.NewReader(sum[:])
	var got [4]uint64
	for i := range hash {
		binary.Read(sumReader, binary.LittleEndian, &got[i])
	}
	if hash != got {
		return ErrInvalidSession
	}

	x.ID = sessionID
	x.Expires = time.Unix(0, int64(expiresAt))
	return nil
}

func (x *Session) String(secret Secret) string {
	str := fmt.Sprintf("%s%016x%016x", secret.String(), x.ID, x.Expires.UnixNano())
	sum := sha256.Sum256([]byte(str))
	sumReader := bytes.NewReader(sum[:])
	var hash [4]uint64
	for i := range hash {
		binary.Read(sumReader, binary.LittleEndian, &hash[i])
	}
	return fmt.Sprintf(SessionFormat,
		hash[0], hash[1], hash[2], hash[3], x.ID, uint64(x.Expires.UnixNano()))
}

func (x *Session) ReadCookie(secret Secret, src *http.Cookie) error {
	return x.ReadString(secret, src.Value)
}

type Secret [4]uint64

var ErrInvalidSecret = errors.New("invalid secret")

func NewSecret() Secret {
	var token Secret
	for i := range token {
		binary.Read(rand.Reader, binary.LittleEndian, &token[i])
	}
	return token
}

func (x Secret) String() string {
	var got [len(x)]string
	for i := range x {
		got[i] = fmt.Sprintf("%016x", x[i])
	}
	return strings.Join(got[:], "-")
}

func DecodeSecret(secretString string) (Secret, error) {
	tokenParts := strings.Split(secretString, "-")
	var token Secret
	if len(tokenParts) != len(token) {
		return Secret{}, ErrInvalidSecret
	}
	for i := range token {
		if len(tokenParts[i]) != 16 {
			return Secret{}, ErrInvalidSecret
		}
		token[i], _ = strconv.ParseUint(tokenParts[i], 16, 64)
	}
	return token, token.Validate()
}

func (x Secret) Validate() error {
	for i := range x {
		if x[i] == 0 {
			return ErrInvalidSecret
		}
	}
	return nil
}
