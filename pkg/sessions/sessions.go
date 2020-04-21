package sessions

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const SessionCookieKey = "vvgo_session"

var ErrSessionNotFound = errors.New("session not found")

type Store struct {
	StoreOpts
	Secret Secret
	cache  *storage.Cache
	locker *locker.Locker
}

type SessionID uint64

type Session struct {
	ID      SessionID
	Expires time.Time
}

type StoreOpts struct {
	CookieName string
	LockerName string
}

const DataFile = "users.json"

type Kind string

func (x Kind) String() string { return string(x) }

const (
	IdentityDiscord Kind = "discord"
)

type Identity struct {
	Kind        `json:"kind"`
	DiscordUser `json:"discord_user,omitempty"`
}

type DiscordUser struct {
	UserID string `json:"user_id"`
}

func NewStore(config StoreOpts) *Store {
	return &Store{
		cache:  storage.NewCache(storage.CacheOpts{}),
		locker: locker.NewLocker(locker.Opts{RedisKey: config.LockerName}),
	}
}

func (x *Store) ReadIdentityFromRequest(ctx context.Context, r *http.Request, dest *Identity) error {
	// read the session
	var session Session
	if err := x.ReadSessionFromRequest(r, &session); err != nil {
		return err
	}
	// lookup the session
	return x.GetIdentity(ctx, session.ID, dest)
}

func (x *Store) GetIdentity(ctx context.Context, id SessionID, dest *Identity) error {
	// read
	if err := x.locker.Lock(ctx); err != nil {
		return err
	}
	defer x.locker.Unlock(ctx)

	// deserialize the data file
	identities := make(map[SessionID]Identity)
	if err := x.getMap(ctx, &identities); err != nil {
		return err
	}

	if _, ok := identities[id]; !ok {
		return ErrSessionNotFound
	}
	*dest = identities[id]
	return nil
}

func (x *Store) getMap(ctx context.Context, dest *map[SessionID]Identity) error {
	// load the data file from cache
	var obj storage.Object
	if err := x.cache.GetObject(ctx, DataFile, &obj); err != nil {
		return err
	}

	if err := json.NewDecoder(&obj.Buffer).Decode(&dest); err != nil {
		return fmt.Errorf("json.Decode() failed: %v", err)
	}
	return nil
}

func (x *Store) ReadSessionFromRequest(r *http.Request, dest *Session) error {
	cookie, err := r.Cookie(SessionCookieKey)
	if err != nil {
		return err
	}
	return dest.ReadCookie(x.Secret, cookie)
}

func (x *Store) StoreIdentity(ctx context.Context, sessionID SessionID, src *Identity) error {
	// read+modify+write
	if err := x.locker.Lock(ctx); err != nil {
		return err
	}
	defer x.locker.Unlock(ctx)

	// read
	var sessions map[SessionID]Identity
	if err := x.getMap(ctx, &sessions); err != nil {
		return err
	}

	// modify
	sessions[sessionID] = *src

	// write
	var obj storage.Object
	if err := json.NewEncoder(&obj.Buffer).Encode(&sessions); err != nil {
		return fmt.Errorf("json.Decode() failed: %v", err)
	}
	if err := x.cache.PutObject(ctx, DataFile, &obj); err != nil {
		return fmt.Errorf("x.cache.PutObject() failed: %v", err)
	}
	return nil
}

func (x *Store) NewCookie(name string, session Session) *http.Cookie {
	return &http.Cookie{
		Name:    name,
		Value:   session.String(x.Secret),
		Expires: session.Expires,
	}
}

func (x *Store) NewSession(expiresAt time.Time) *Session {
	var id SessionID
	binary.Read(rand.Reader, binary.LittleEndian, &id)
	return &Session{
		ID:      id,
		Expires: expiresAt,
	}
}

var ErrInvalidSession = errors.New("invalid cookie")

const SessionFormat = "%016x-V-%016x-V-%016x-G-%016x-O-%016x%016x"

func (x *Session) ReadCookie(secret Secret, src *http.Cookie) error {
	return x.ReadString(secret, src.Value)
}

func (x *Session) ReadString(secret Secret, value string) error {
	// read the cookie
	var hash [4]uint64
	var id SessionID
	var expiresAt uint64
	_, err := fmt.Sscanf(value, SessionFormat,
		&hash[0], &hash[1], &hash[2], &hash[3], &id, &expiresAt)
	if err != nil {
		return ErrInvalidSession
	}

	// validate the cookie
	str := fmt.Sprintf("%s%016x%016x", secret.String(), id, expiresAt)
	sum := sha256.Sum256([]byte(str))
	sumReader := bytes.NewReader(sum[:])
	var got [4]uint64
	for i := range hash {
		binary.Read(sumReader, binary.LittleEndian, &got[i])
	}
	if hash != got {
		return ErrInvalidSession
	}

	x.ID = id
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
