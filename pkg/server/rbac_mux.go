package server

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
	"net/url"
	"strings"
)

// RBACMux Authenticate http requests using session based authentication.
// If the request has a valid session or token with the required role, it is allowed access.
type RBACMux struct {
	Basic  map[[2]string][]models.Role
	Bearer map[string][]models.Role
	*http.ServeMux
}

// HandleFunc registers the handler function for the given pattern.
func (auth *RBACMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request), role models.Role) {
	auth.Handle(pattern, http.HandlerFunc(handler), role)
}

func (auth *RBACMux) Handle(pattern string, handler http.Handler, role models.Role) {
	auth.ServeMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var identity models.Identity
		switch {
		case auth.readBasicAuth(r, &identity):
			break
		case auth.readBearer(r, &identity):
			break
		case auth.readSession(ctx, r, &identity):
			break
		default:
			identity = models.Anonymous()
		}

		if values := r.URL.Query(); len(values["roles"]) != 0 {
			wantRoles := make([]models.Role, len(values["roles"]))
			for i := range values["roles"] {
				wantRoles[i] = models.Role(values["roles"][i])
			}
			identity = identity.AssumeRoles(wantRoles...)
		}

		if identity.HasRole(role) {
			if role != models.RoleAnonymous {
				logger.WithField("roles", identity.Roles).WithField("path", r.URL.Path).Info("access granted")
			}
			handler.ServeHTTP(w, r.Clone(context.WithValue(ctx, login.CtxKeyVVGOIdentity, &identity)))
			return
		}
		logger.WithField("roles", identity.Roles).WithField("path", r.URL.Path).Info("access denied")

		if identity.IsAnonymous() && strings.HasPrefix(r.URL.Path, "/api") == false {
			values := make(url.Values)
			values.Set("target", r.RequestURI)
			http.Redirect(w, r, "/login?"+values.Encode(), http.StatusFound)
			return
		}
		http_helpers.Unauthorized(ctx, w)
	})
}

func (auth *RBACMux) readBasicAuth(r *http.Request, dest *models.Identity) bool {
	if auth.Basic == nil {
		return false
	}

	user, pass, _ := r.BasicAuth()
	gotRoles, ok := auth.Basic[[2]string{user, pass}]
	if !ok {
		return false
	}

	*dest = models.Identity{
		Kind:  models.KindBasic,
		Roles: gotRoles,
	}
	return true
}

func (auth *RBACMux) readBearer(r *http.Request, dest *models.Identity) bool {
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

	*dest = models.Identity{
		Kind:  models.KindBearer,
		Roles: gotRoles,
	}
	return true
}

func (auth *RBACMux) readSession(ctx context.Context, r *http.Request, identity *models.Identity) bool {
	return login.ReadSessionFromRequest(ctx, r, identity) == nil
}
