package sessions

import (
	"context"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"net/http"
)

const SessionCookieKey = "vvgo_session"

var ErrSessionNotFound = errors.New("session not found")

type Store struct {
	Opts
	sessions map[string]Session
	locker   *locker.Locker
}

type Opts struct {
	LockerName string
}

type Session struct {
	DiscordUser  string
	DiscordRoles []string
}

func NewStore(config Opts) *Store {
	return &Store{
		sessions: make(map[string]Session),
		locker:   locker.NewLocker(locker.Opts{RedisKey: config.LockerName}),
	}
}

func (x *Store) Add(ctx context.Context, key string, session *Session) error {
	if err := x.locker.Lock(ctx); err != nil {
		return fmt.Errorf("x.locker.Lock() failed: %v", err)
	}
	defer x.locker.Unlock(ctx)
	x.sessions[key] = *session
	return nil
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
