package server

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"net/http"
	"net/url"
	"strings"
)

// RBACMux Authenticate http requests using the sessions api.
// If the request has a valid session or token with the required role, it is allowed access.
type RBACMux struct {
	Basic  map[[2]string][]login.Role
	Bearer map[string][]login.Role
	*http.ServeMux
}

// HandleFunc registers the handler function for the given pattern.
func (auth *RBACMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request), role login.Role) {
	auth.Handle(pattern, http.HandlerFunc(handler), role)
}

func (auth *RBACMux) Handle(pattern string, handler http.Handler, role login.Role) {
	auth.ServeMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var identity login.Identity
		switch {
		case auth.readBasicAuth(r, &identity):
			break
		case auth.readBearer(r, &identity):
			break
		case auth.readSession(ctx, r, &identity):
			break
		default:
			identity = login.Anonymous()
		}

		if values := r.URL.Query(); len(values["roles"]) != 0 {
			wantRoles := make([]login.Role, len(values["roles"]))
			for i := range values["roles"] {
				wantRoles[i] = login.Role(values["roles"][i])
			}
			identity = identity.AssumeRoles(wantRoles...)
		}

		if identity.HasRole(role) {
			if role != login.RoleAnonymous {
				logger.WithField("roles", identity.Roles).WithField("path", r.URL.Path).Info("access granted")
			}
			handler.ServeHTTP(w, r.Clone(context.WithValue(ctx, login.CtxKeyVVGOIdentity, &identity)))
			return
		}
		logger.WithField("roles", identity.Roles).WithField("path", r.URL.Path).Info("access denied")

		if identity.IsAnonymous() {
			values := make(url.Values)
			values.Set("target", r.RequestURI)
			http.Redirect(w, r, "/login?"+values.Encode(), http.StatusFound)
			return
		}
		helpers.Unauthorized(w)
	})
}

func (auth *RBACMux) readBasicAuth(r *http.Request, dest *login.Identity) bool {
	if auth.Basic == nil {
		return false
	}

	user, pass, _ := r.BasicAuth()
	gotRoles, ok := auth.Basic[[2]string{user, pass}]
	if !ok {
		return false
	}

	*dest = login.Identity{
		Kind:  login.KindBasic,
		Roles: gotRoles,
	}
	return true
}

func (auth *RBACMux) readBearer(r *http.Request, dest *login.Identity) bool {
	if auth.Bearer == nil {
		return false
	}

	bearer := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(bearer, "Bearer ") {
		return false
	}
	bearer = bearer[len("Bearer "):]

	gotRoles, ok := auth.Bearer[bearer]
	if !ok {
		return false
	}

	*dest = login.Identity{
		Kind:  login.KindBearer,
		Roles: gotRoles,
	}
	return true
}

func (auth *RBACMux) readSession(ctx context.Context, r *http.Request, identity *login.Identity) bool {
	return login.ReadSessionFromRequest(ctx, r, identity) == nil
}
