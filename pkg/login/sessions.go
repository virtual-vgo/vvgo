package login

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
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

// Store provides access to the map of session id's to access roles.
// It can read and validate signed session cookies from incoming requests,
// and create new signed cookies for authenticated users.
type Store struct {
	config Config
	cache  *storage.Cache
	locker *locker.Locker
}

type Config struct {
	// Secret is the secret used to sign the session data
	Secret Secret `default:"0000000000000000000000000000000000000000000000000000000000000000"`

	// CookieName is the name of the cookies created by the store.
	CookieName string `split_words:"true" default:"vvgo-sessions"`

	// CookieDomain is the domain where the cookies can be used.
	// This should be the domain that users visit in their browser.
	CookieDomain string `split_words:"true" default:"localhost"`

	// CookiePath is the url path where the cookies can be used.
	CookiePath string `split_words:"true" default:"/"`

	// RedisKey is the key name used when obtaining locks in redis.
	RedisKey string `split_words:"true"`
}

const DataFile = "users.json"

// NewStore returns a new sessions client.
func NewStore(locksmith *locker.Locksmith, config Config) *Store {
	return &Store{
		config: config,
		cache:  storage.NewCache(storage.CacheOpts{}),
		locker: locksmith.NewLocker(locker.Opts{RedisKey: config.RedisKey}),
	}
}

// Init initializes the storage map so that it is ready for use.
func (x *Store) Init(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx, "sessions_store_init")
	defer span.Send()
	obj := storage.NewJSONObject([]byte(`{}`))
	if err := x.cache.PutObject(ctx, DataFile, obj); err != nil {
		return fmt.Errorf("x.cache.PutObject() failed: %w", err)
	}
	return nil
}

// ReadIdentityFromRequest reads the identity from the sessions db based on the request data.
func (x *Store) ReadIdentityFromRequest(ctx context.Context, r *http.Request, dest *Identity) error {
	var session Session
	if err := x.ReadSessionFromRequest(r, &session); err != nil {
		return err
	}
	return x.GetIdentity(ctx, session.ID, dest)
}

// Reads the session data from an http request.
func (x *Store) ReadSessionFromRequest(r *http.Request, dest *Session) error {
	cookie, err := r.Cookie(x.config.CookieName)
	if err == nil {
		return dest.DecodeCookie(x.config.Secret, cookie)
	}

	return ErrSessionNotFound
}

// NewCookie returns cookie with a cryptographically signed session payload.
func (x *Store) NewCookie(session Session) *http.Cookie {
	return &http.Cookie{
		Name:     x.config.CookieName,
		Value:    session.SignAndEncode(x.config.Secret),
		Expires:  time.Unix(0, int64(session.Expires)),
		Domain:   x.config.CookieDomain,
		Path:     x.config.CookiePath,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
	}
}

// NewSession returns a new session with a crypto-rand session id.
func (x *Store) NewSession(expiresAt time.Time) Session {
	var id SessionID
	binary.Read(rand.Reader, binary.LittleEndian, &id)
	return Session{
		ID:      id,
		Expires: uint64(expiresAt.Unix()),
	}
}

// GetIdentity reads the login identity for the given session ID.
func (x *Store) GetIdentity(ctx context.Context, id SessionID, dest *Identity) error {
	if err := x.locker.Lock(ctx); err != nil {
		return err
	}
	defer x.locker.Unlock(ctx)

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

// StoreIdentity stores the session id and identity.
func (x *Store) StoreIdentity(ctx context.Context, id SessionID, src *Identity) error {
	if err := x.locker.Lock(ctx); err != nil {
		return err
	}
	defer x.locker.Unlock(ctx)

	var sessions map[SessionID]Identity
	if err := x.getMap(ctx, &sessions); err != nil {
		return err
	}

	sessions[id] = *src
	return x.writeMap(ctx, &sessions)
}

// DeleteIdentity deletes the sessionID key from the map.
func (x *Store) DeleteIdentity(ctx context.Context, id SessionID) error {
	if err := x.locker.Lock(ctx); err != nil {
		return err
	}
	defer x.locker.Unlock(ctx)

	var sessions map[SessionID]Identity
	if err := x.getMap(ctx, &sessions); err != nil {
		return err
	}

	delete(sessions, id)

	return x.writeMap(ctx, &sessions)
}

func (x *Store) getMap(ctx context.Context, dest *map[SessionID]Identity) error {
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

var ErrInvalidSession = errors.New("invalid session")
var ErrInvalidSignature = errors.New("invalid signature")

const SessionFormat = "V-i-r-t-u-a-l--V-G-O--%016x%016x%016x%016x%016x%016x"

// Sessions can be encoded as url safe strings and embedded into cookies or http headers.
// Session id's are used look up user access roles.
// Each session should get a unique id.
type Session struct {
	// Id is a unique random session id
	ID SessionID

	// Expires is the time in second since epoch that this session expires
	Expires uint64
}

type SessionID uint64

// SignAndEncode signs and encodes the session into a url safe string.
// The secret is used to cryptographically sign the session data.
func (x *Session) SignAndEncode(secret Secret) string {
	hash := x.makeSignature(secret)
	return fmt.Sprintf(SessionFormat, hash[0], hash[1], hash[2], hash[3], x.ID, x.Expires)
}

// DecodeCookie reads session data stored in the cookie.
// Secret is used to validate the cookie's signature.
func (x *Session) DecodeCookie(secret Secret, src *http.Cookie) error {
	return x.DecodeAndValidate(secret, src.Value)
}

// DecodeAndValidate reads session from a string and validates the signature.
// Secret is used to validate the session signature.
// * ErrInvalidSession is returned when the session code not be decoded.
// * ErrInvalidSignature is returned when the session was read successfully, but signature was invalid.
// * ErrSessionExpired is returned when the session was read and has a valid signature, but the session is expired.
// If ErrInvalidSignature or ErrSessionExpired, the session data is still written to this object.
func (x *Session) DecodeAndValidate(secret Secret, value string) error {
	// read the cookie
	var sig [4]uint64
	_, err := fmt.Sscanf(value, SessionFormat,
		&sig[0], &sig[1], &sig[2], &sig[3], &x.ID, &x.Expires)
	if err != nil {
		return ErrInvalidSession
	}

	switch {
	case x.makeSignature(secret) != sig:
		return ErrInvalidSignature
	case uint64(time.Now().Unix()) >= x.Expires:
		return ErrSessionExpired
	default:
		return nil
	}
}

func (x *Session) makeSignature(secret Secret) [4]uint64 {
	str := fmt.Sprintf("%v|%v|%v", secret, x.ID, x.Expires)
	sum := sha256.Sum256([]byte(str))
	sumReader := bytes.NewReader(sum[:])
	var hash [4]uint64
	for i := range hash {
		binary.Read(sumReader, binary.LittleEndian, &hash[i])
	}
	return hash
}

var ErrSessionExpired = errors.New("session is expired")

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
