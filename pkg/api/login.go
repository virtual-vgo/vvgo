package api

import (
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/access"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
)

type PasswordLoginHandler struct {
	Sessions *sessions.Store
	Logins   []PasswordLogin
}

type PasswordLogin struct {
	User  string
	Pass  string
	Roles []access.Role
}

func (x PasswordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "password_login")
	defer span.Send()

	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
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
		Kind:  sessions.KindPassword,
		Roles: roles,
	}
	loginRedirect(ctx, w, r, x.Sessions, &identity)
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
