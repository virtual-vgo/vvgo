package sessions

import (
	"context"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"net/http"
)

const SessionKey = "vvgo_session"

var sessions *Sessions

func init() {
	sessions = &Sessions{
		sessions: make(map[string]Session),
		locker:   locker.NewLocker(locker.Opts{RedisKey: SessionKey}),
	}
}

func ReadFromRequest(ctx context.Context, r *http.Request, dest *Session) error {
	return sessions.ReadFromRequest(ctx, r, dest)
}

func Add(ctx context.Context, session *Session) error {
	return sessions.Add(ctx, session)
}

type Config struct {
	LockerName string
}

type Sessions struct {
	sessions map[string]Session
	locker   *locker.Locker
}

type Session struct {
	Key       string
	VVVGOUser string
}

func (x *Sessions) Add(ctx context.Context, session *Session) error {
	if err := x.locker.Lock(ctx); err != nil {
		return fmt.Errorf("x.locker.Lock() failed: %v", err)
	}
	defer x.locker.Unlock(ctx)
	x.sessions[session.Key] = *session
	return nil
}

var ErrSessionNotFound = errors.New("session not found")

func (x *Sessions) Get(ctx context.Context, key string, dest *Session) error {
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

func (x *Sessions) ReadFromRequest(ctx context.Context, r *http.Request, dest *Session) error {
	cookie, err := r.Cookie(SessionKey)
	if err != nil {
		return err
	}
	return sessions.Get(ctx, cookie.Value, dest)
}
