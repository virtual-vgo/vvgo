package api

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/access"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"html/template"
	"net/http"
	"path/filepath"
)

type Login struct {
	User  string
	Pass  string
	Roles []access.Role
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
		var identity sessions.Identity
		if err := x.Sessions.ReadIdentityFromRequest(ctx, r, &identity); err == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		opts := x.NavBar.NewOpts(ctx, r)
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
		var identity sessions.Identity
		if err := x.Sessions.ReadIdentityFromRequest(ctx, r, &identity); err == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		user := r.FormValue("user")
		pass := r.FormValue("pass")

		var roles []access.Role
		for _, login := range x.Logins {
			if user == login.User && pass == login.Pass {
				roles = login.Roles
			}
		}

		if len(roles) == 0 {
			logger.WithFields(logrus.Fields{
				"user": user,
			}).Error("password authentication failed")
			unauthorized(w)
			return
		}

		// create the identity object
		identity = sessions.Identity{
			Kind:  sessions.IdentityPassword,
			Roles: roles,
		}
		loginRedirect(newCookie(ctx, x.Sessions, &identity), w, r, "/")

	default:
		methodNotAllowed(w)
	}
}

type LogoutHandler struct {
	Sessions *sessions.Store
}

func (x LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "logout_handler")
	defer span.Send()

	var session sessions.Session
	if err := x.Sessions.ReadSessionFromRequest(r, &session); err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// delete the session
	if err := x.Sessions.DeleteIdentity(ctx, session.ID); err != nil {
		logger.WithError(err).Error("x.Sessions.DeleteIdentity() failed")
		internalServerError(w)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
	return
}
