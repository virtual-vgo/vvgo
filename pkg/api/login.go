package api

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type Login struct {
	User  string
	Pass  string
	Roles []string
}

type LoginHandler struct {
	NavBar   NavBar
	Secret   sessions.Secret
	Sessions *sessions.Store
	Logins   []Login
}

func (x LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "login_handler")
	defer span.Send()

	switch r.Method {
	case http.MethodGet:
		var session sessions.Session
		err := x.Sessions.ReadSessionFromRequest(r, &session);
		if err == nil {
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
		if err := x.Sessions.ReadSessionFromRequest(r, &session); err == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		user := r.FormValue("user")
		pass := r.FormValue("pass")

		var roles []string
		for _, login := range x.Logins {
			if user == login.User && pass == login.Pass {
				roles = login.Roles
			}
		}

		if len(roles) == 0 {
			logger.WithFields(logrus.Fields{
				"user": user,
				"pass": pass,
			}).Error("authorization failed")
			unauthorized(w)
			return
		}

		// create the identity object
		identity := sessions.Identity{
			Kind:  sessions.IdentityPassword,
			Roles: roles,
		}

		// create a session and cookie
		session = x.Sessions.NewSession(time.Now().Add(7 * 24 * 3600 * time.Second))
		cookie := x.Sessions.NewCookie(session)
		if err := x.Sessions.StoreIdentity(ctx, session.ID, &identity); err != nil {
			logger.WithError(err).Error("x.Sessions.StoreIdentity() failed")
			internalServerError(w)
			return
		}

		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusFound)

	default:
		methodNotAllowed(w)
	}
}
