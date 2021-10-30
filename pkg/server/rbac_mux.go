package server

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

// RBACMux Authenticate http requests using session based authentication.
// If the request has a valid session or token with the required role, it is allowed access.
type RBACMux struct {
	*http.ServeMux
}

// HandleFunc registers the handler function for the given pattern.
func (auth *RBACMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request), role models.Role) {
	auth.Handle(pattern, http.HandlerFunc(handler), role)
}

// HandleApiFunc registers the handler function for the given pattern.
func (auth *RBACMux) HandleApiFunc(pattern string, handler func(*http.Request) models.ApiResponse, role models.Role) {
	auth.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		resp := handler(r)

		switch resp.Status {
		case models.StatusFound:
			http.Redirect(w, r, resp.Location, http.StatusFound)

		case models.StatusError:
			if resp.Error != nil {
				w.WriteHeader(resp.Error.Code)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.JsonEncodeFailure(ctx, err)
		}
	}), role)
}

func (auth *RBACMux) Handle(pattern string, handler http.Handler, role models.Role) {
	auth.ServeMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var identity models.Identity
		login.ReadSessionFromRequest(ctx, r, &identity)

		if identity.HasRole(role) {
			if role != models.RoleAnonymous {
				logger.WithField("roles", identity.Roles).WithField("path", r.URL.Path).Info("access granted")
			}
			handler.ServeHTTP(w, r.Clone(context.WithValue(ctx, login.CtxKeyVVGOIdentity, &identity)))
			return
		}
		logger.WithField("roles", identity.Roles).WithField("path", r.URL.Path).Info("access denied")
		http_helpers.WriteUnauthorizedError(ctx, w)
	})
}
