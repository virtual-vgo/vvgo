package sessions

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"net/http"
)

type Locker interface {
	Lock(ctx context.Context) bool
	Unlock()
}

const SessionCookie = "vvgo_session"

var logger = log.Logger()
var sessions = new(Sessions)

func Add(ctx context.Context, session Session)                  { sessions.Add(ctx, session) }
func Read(ctx context.Context, r *http.Request) (Session, bool) { return sessions.Read(ctx, r) }

type Sessions struct {
	sessions []Session
	locker   Locker
}

func (x *Sessions) Add(ctx context.Context, session Session) {
	x.locker.Lock(ctx)
	defer x.locker.Unlock()
	x.sessions = append(x.sessions, session)
}

func (x *Sessions) Read(ctx context.Context, r *http.Request) (Session, bool) {
	x.locker.Lock(ctx)
	defer x.locker.Unlock()
	cookie, err := r.Cookie(SessionCookie)
	if err != nil {
		logger.WithError(err).Debug("cookie error")
		return Session{}, false
	}
	return sessions.Get(cookie.Value)
}

func (x *Sessions) Get(key string) (Session, bool) {
	for _, session := range x.sessions {
		if session.Key == key {
			return session, true
		}
	}
	return Session{}, false
}

type Session struct {
	Key       string
	VVVGOUser string
}
