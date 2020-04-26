package access

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
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"strings"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

type Store struct {
	config Config
	cache  *storage.Cache
	locker *locker.Locker
}

type SessionID uint64

type Session struct {
	ID      SessionID
	Expires time.Time
}

type Config struct {
	Secret       Secret `default:"0000000000000000000000000000000000000000000000000000000000000000"`
	CookieName   string `split_words:"true" default:"vvgo-sessions"`
	CookieDomain string `split_words:"true" default:"localhost"`
	CookiePath   string `split_words:"true" default:"/"`
	RedisKey     string `split_words:"true"`
}

const DataFile = "users.json"

// NewStore returns a new sessions client.
func NewStore(lockSmith *locker.LockSmith, config Config) *Store {
	return &Store{
		config: config,
		cache:  storage.NewCache(storage.CacheOpts{}),
		locker: lockSmith.NewLocker(locker.Opts{RedisKey: config.RedisKey}),
	}
}

// Init initializes the storage map so that it is ready for use.
func (x *Store) Init(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx, "sessions_store_init")
	defer span.Send()
	obj := storage.Object{Bytes: []byte(`{}`)}
	if err := x.cache.PutObject(ctx, DataFile, &obj); err != nil {
		return fmt.Errorf("x.cache.PutObject() failed: %v", err)
	}
	return nil
}

// ReadIdentityFromRequest reads the identity from the sessions db based on the request data.
func (x *Store) ReadIdentityFromRequest(ctx context.Context, r *http.Request, dest *Identity) error {
	// read the session
	var session Session
	if err := x.ReadSessionFromRequest(r, &session); err != nil {
		return err
	}
	// lookup the session
	return x.GetIdentity(ctx, session.ID, dest)
}

// Reads the session data from an http request
// Currently, this function can read sessions from either a cookie or bearer token.
func (x *Store) ReadSessionFromRequest(r *http.Request, dest *Session) error {
	// check for a bearer token
	token := r.Header.Get("Authorization")
	if strings.HasPrefix(token, "Bearer ") {
		return dest.Decode(x.config.Secret, strings.TrimPrefix(token, "Bearer "))
	}

	// check for a cookie
	cookie, err := r.Cookie(x.config.CookieName)
	if err == nil {
		return dest.DecodeCookie(x.config.Secret, cookie)
	}

	return ErrSessionNotFound
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

func (x *Store) StoreIdentity(ctx context.Context, id SessionID, src *Identity) error {
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
	sessions[id] = *src

	// write
	return x.writeMap(ctx, &sessions)
}

func (x *Store) DeleteIdentity(ctx context.Context, id SessionID) error {
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
	delete(sessions, id)

	// write
	return x.writeMap(ctx, &sessions)
}

func (x *Store) NewCookie(session Session) *http.Cookie {
	return &http.Cookie{
		Name:     x.config.CookieName,
		Value:    session.Encode(x.config.Secret),
		Expires:  session.Expires,
		Domain:   x.config.CookieDomain,
		Path:     x.config.CookiePath,
		HttpOnly: true,
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

func (x *Store) getMap(ctx context.Context, dest *map[SessionID]Identity) error {
	// load the data file from cache
	var obj storage.Object
	if err := x.cache.GetObject(ctx, DataFile, &obj); err != nil {
		return err
	}

	if err := json.NewDecoder(bytes.NewReader(obj.Bytes)).Decode(&dest); err != nil {
		return fmt.Errorf("json.Decode() failed: %v", err)
	}
	return nil
}

func (x *Store) writeMap(ctx context.Context, src *map[SessionID]Identity) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(src); err != nil {
		return fmt.Errorf("json.Decode() failed: %v", err)
	}
	if err := x.cache.PutObject(ctx, DataFile, storage.NewJSONObject(buf.Bytes())); err != nil {
		return fmt.Errorf("x.cache.PutObject() failed: %v", err)
	}
	return nil
}

var ErrInvalidSession = errors.New("invalid cookie")

const SessionFormat = "V-i-r-t-u-a-l--V-G-O--%016x%016x%016x%016x%016x%016x"

func (x *Session) Encode(secret Secret) string {
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

func (x *Session) DecodeCookie(secret Secret, src *http.Cookie) error {
	return x.Decode(secret, src.Value)
}

func (x *Session) Decode(secret Secret, value string) error {
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
	return x.Validate()
}

var ErrSessionExpired = errors.New("session is expired")

func (x *Session) Validate() error {
	switch {
	case time.Now().UnixNano() >= x.Expires.UnixNano():
		return ErrSessionExpired
	default:
		return nil
	}
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

func (x *Secret) Decode(src string) error {
	_, err := fmt.Sscanf(src, SecretFormat, &x[0], &x[1], &x[2], &x[3])
	return err
}
