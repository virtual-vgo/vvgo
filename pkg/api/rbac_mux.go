package api

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
)

// RBACMux Authenticate http requests using session based authentication.
// If the request has a valid session or token with the required role, it is allowed access.
type RBACMux struct {
	*http.ServeMux
}

func NewRBACMux() RBACMux { return RBACMux{ServeMux: http.NewServeMux()} }

// HandleFunc registers the handler function for the given pattern.
func (mux *RBACMux) HandleFunc(pattern string, handler http.HandlerFunc, role auth.Role) {
	mux.Handle(pattern, handler, role)
}

// HandleApiFunc registers the handler function for the given pattern.
func (mux *RBACMux) HandleApiFunc(pattern string, handler func(*http.Request) Response, role auth.Role) {
	mux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(r).WriteHTTP(r.Context(), w, r)
	}), role)
}

func (mux *RBACMux) Handle(pattern string, handler http.Handler, role auth.Role) {
	mux.ServeMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var identity auth.Identity
		auth.ReadIdentityFromRequest(ctx, r, &identity)

		if identity.HasRole(role) {
			if role != auth.RoleAnonymous {
				logger.WithField("roles", identity.Roles).WithField("path", r.URL.Path).Info("http server: access granted")
			}
			handler.ServeHTTP(w, r.Clone(context.WithValue(ctx, auth.CtxKeyVVGOIdentity, &identity)))
			return
		} else {
			logger.WithField("roles", identity.Roles).WithField("path", r.URL.Path).Info("http server: access denied")
			response.NewUnauthorizedError().WriteHTTP(ctx, w, r)
		}
	})
}
