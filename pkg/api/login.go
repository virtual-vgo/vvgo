package api

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type LoginHandler struct {
	NavBar
}

const SessionCookie = "vvgo_session"

var sessions = new(Sessions)

type Sessions struct {
	sessions []Session
}

func (x *Sessions) Add(session Session) {
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

func (x LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, ok := sessions.Read(r)
	if !ok {
		user := r.FormValue("user")
		pass := r.FormValue("pass")
		if user == "jackson" && pass == "jackson" {
			cookie := http.Cookie{
				Name:    SessionCookie,
				Value:   user,
				Expires: time.Now().Add(3600 * time.Second),
			}
			http.SetCookie(w, &cookie)
			sessions.Add(Session{
				Key:       user,
				VVVGOUser: user,
			})
			w.Write([]byte("welcome jackson, have a cookie!\n"))
		} else {
			unauthorized(w)
		}
	} else {
		fmt.Fprint(w, "welcome back "+session.VVVGOUser)
	}
}

func (x LoginHandler) __ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		opts := x.NavBar.NewOpts(r)
		page := struct {
			Header template.HTML
			NavBar template.HTML
		}{
			Header: Header(),
			NavBar: x.NavBar.RenderHTML(opts),
		}

		var buf bytes.Buffer
		if ok := parseAndExecute(&buf, &page, filepath.Join(PublicFiles, "login.gohtml")); !ok {
			internalServerError(w)
		} else {
			buf.WriteTo(w)
		}
	case http.MethodPost:
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
