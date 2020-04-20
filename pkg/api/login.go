package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type LoginHandler struct {
	NavBar   NavBar
	Sessions *sessions.Store
}

func (x LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "login_handler")
	defer span.Send()

	switch r.Method {
	case http.MethodGet:
		var session sessions.Session
		if err := x.Sessions.ReadFromRequest(ctx, r, &session); err == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

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
			return
		}
		buf.WriteTo(w)

	case http.MethodPost:
		var session sessions.Session
		if err := x.Sessions.ReadFromRequest(ctx, r, &session); err == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		user := r.FormValue("user")
		pass := r.FormValue("pass")
		if !(user == "jackson@jacksonargo.com" || pass == "jackson") {
			logger.WithFields(logrus.Fields{
				"user": user,
				"pass": pass,
			}).Error("authorization failed")
			unauthorized(w)
			return
		}

		token := ""
		if err := x.Sessions.Add(ctx, token, &sessions.Session{
		}); err != nil {
			tracing.AddError(ctx, err)
			logger.WithError(err).Error("x.Sessions.Add() failed")
		}

		cookie := http.Cookie{
			Name:    sessions.SessionCookieKey,
			Value:   token,
			Expires: time.Now().Add(3600 * time.Second),
		}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusFound)

	default:
		methodNotAllowed(w)
	}
}

type Session struct {
	ID      uint64
	Expires time.Time
}

var ErrInvalidSession = errors.New("invalid cookie")

const SessionFormat = "%016x-The-%016x-Earth-%016x-Is-%016x-Flat-%016x%016x"

func (x *Session) ReadString(secret Secret, value string) error {
	// read the cookie
	var hash [4]uint64
	var sessionID uint64
	var expiresAt uint64
	_, err := fmt.Sscanf(value, SessionFormat,
		&hash[0], &hash[1], &hash[2], &hash[3], &sessionID, &expiresAt)
	if err != nil {
		return ErrInvalidSession
	}

	// validate the cookie
	str := fmt.Sprintf("%s%016x%016x", secret.String(), sessionID, expiresAt)
	sum := sha256.Sum256([]byte(str))
	sumReader := bytes.NewReader(sum[:])
	var got [4]uint64
	for i := range hash {
		binary.Read(sumReader, binary.LittleEndian, &got[i])
	}
	if hash != got {
		return ErrInvalidSession
	}

	x.ID = sessionID
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

func (x *Session) ReadCookie(secret Secret, src *http.Cookie) error {
	return x.ReadString(secret, src.Value)
}

func (x *Session) RenderCookie(secret Secret, dest *http.Cookie) {
	*dest = http.Cookie{
		Value:   x.String(secret),
		Expires: x.Expires,
	}
	return
}
