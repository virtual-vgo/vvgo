package sessions

import (
	"context"
	"net/http"
)

type Locker interface {
	Lock(ctx context.Context) bool
	Unlock()
}

const SessionCookie = "vvgo_session"

var sessions = new(Sessions)

func Add(session Session)                  { sessions.Add(session) }
func Read(r *http.Request) (Session, bool) { return sessions.Read(r) }

type Sessions struct {
	sessions []Session
	locker   Locker
}

func (x *Sessions) Add(ctx context.Context, session Session) {
	x.locker.Lock()
	defer x.locker.Unlock()
	x.sessions = append(x.sessions, session)
}

func (x *Sessions) Read(r *http.Request) (Session, bool) {
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
