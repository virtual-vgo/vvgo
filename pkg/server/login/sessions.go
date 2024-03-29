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
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")

const CtxKeyVVGOIdentity = "vvgo_identity"

func IdentityFromContext(ctx context.Context) models.Identity {
	ctxIdentity := ctx.Value(CtxKeyVVGOIdentity)
	identity, ok := ctxIdentity.(*models.Identity)
	if !ok {
		identity = new(models.Identity)
		*identity = models.Anonymous()
	}
	return *identity
}

// ReadSessionFromRequest reads the identity from the sessions db based on the request data.
func ReadSessionFromRequest(ctx context.Context, r *http.Request, dest *models.Identity) {
	bearer := strings.TrimSpace(r.Header.Get("Authorization"))
	token := r.URL.Query().Get("token")

	var err error
	switch {
	case strings.HasPrefix(bearer, "Bearer "):
		err = GetSession(ctx, bearer[len("Bearer "):], dest)
	case token != "":
		err = GetSession(ctx, token, dest)
	default:
		*dest = models.Anonymous()
	}

	if err != nil {
		logger.MethodFailure(ctx, "login.GetSession", err)
	}
}

func NewSessionKey() string {
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
	identity.Key = NewSessionKey()
	expiresAt := time.Now().Add(expires)
	identity.ExpiresAt = expiresAt
	identity.CreatedAt = time.Now()
	key := "sessions:" + identity.Key
	stringExpires := strconv.Itoa(int(expires.Seconds()))
	srcBytes, _ := json.Marshal(identity)
	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", key, stringExpires, string(srcBytes))); err != nil {
		return "", err
	}
	return identity.Key, nil
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
