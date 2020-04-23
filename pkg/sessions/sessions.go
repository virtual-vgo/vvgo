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
	"github.com/virtual-vgo/vvgo/pkg/access"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

type Store struct {
	Config
	secret Secret
	cache  *storage.Cache
	locker *locker.Locker
}

type SessionID uint64

type Session struct {
	ID      SessionID
	Expires time.Time
}

type Config struct {
	CookieName string
	LockerName string
}

const DataFile = "users.json"

type Kind string

func (x Kind) String() string { return string(x) }

const (
	IdentityPassword Kind = "password"
	IdentityDiscord  Kind = "discord"
)

type Identity struct {
	Kind        `json:"kind"`
	Roles       []access.Role `roles:"roles"`
	DiscordUser `json:"discord_user,omitempty"`
}

type DiscordUser struct {
	UserID string `json:"user_id"`
}

func NewStore(secret Secret, config Config) *Store {
	return &Store{
		Config: config,
		secret: secret,
		cache:  storage.NewCache(storage.CacheOpts{}),
		locker: locker.NewLocker(locker.Opts{RedisKey: config.LockerName}),
	}
}

func (x *Store) Init(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx, "sessions_store_init")
	defer span.Send()
	obj := storage.Object{Buffer: *bytes.NewBuffer([]byte(`{}`))}
	if err := x.cache.PutObject(ctx, DataFile, &obj); err != nil {
		return fmt.Errorf("x.cache.PutObject() failed: %v", err)
	}
	return nil
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
	cookie, err := r.Cookie(x.CookieName)
	if err != nil {
		return err
	}
	return dest.ReadCookie(x.secret, cookie)
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

func (x *Store) NewCookie(session Session) *http.Cookie {
	return &http.Cookie{
		Name:    x.Config.CookieName,
		Value:   session.String(x.secret),
		Expires: session.Expires,
	}
}

func (x *Store) NewSession(expiresAt time.Time) Session {
	var id SessionID
	binary.Read(rand.Reader, binary.LittleEndian, &id)
	return Session{
		ID:      id,
		Expires: expiresAt,
	}
}

var ErrInvalidSession = errors.New("invalid cookie")

const SessionFormat = "V-i-r-t-u-a-l--V-G-O--%016x%016x%016x%016x%016x%016x"

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

const SecretFormat = "%016x%016x%016x%016x"

var ErrInvalidSecret = errors.New("invalid secret")

func NewSecret() Secret {
	var token Secret
	for i := range token {
		binary.Read(rand.Reader, binary.LittleEndian, &token[i])
	}
	return token
}

func (x Secret) Validate() error {
	for i := range x {
		if x[i] == 0 {
			return ErrInvalidSecret
		}
	}
	return nil
}

func (x Secret) String() string {
	return fmt.Sprintf(SecretFormat, x[0], x[1], x[2], x[3])
}

func (x *Secret) Decode(str string) error {
	_, err := fmt.Sscanf(str, SecretFormat, &x[0], &x[1], &x[2], &x[3])
	return err
}
