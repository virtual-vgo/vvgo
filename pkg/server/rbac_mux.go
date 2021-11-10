package server

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

// RBACMux Authenticate http requests using session based authentication.
// If the request has a valid session or token with the required role, it is allowed access.
type RBACMux struct {
	*mux.Router
}

func NewRBACMux() RBACMux { return RBACMux{Router: mux.NewRouter()} }

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
			return

		case models.StatusError:
			w.Header().Set("Content-Type", "application/json")
			if resp.Error != nil {
				w.WriteHeader(resp.Error.Code)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

		case models.StatusOk:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.JsonEncodeFailure(ctx, err)
		}

	}), role)
}

func (auth *RBACMux) Handle(pattern string, handler http.Handler, role models.Role) {
	auth.Router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var identity models.Identity
		login.ReadSessionFromRequest(ctx, r, &identity)

		if identity.HasRole(role) {
			if role != models.RoleAnonymous {
				logger.WithField("roles", identity.Roles).WithField("path", r.URL.Path).Info("http server: access granted")
			}
			handler.ServeHTTP(w, r.Clone(context.WithValue(ctx, login.CtxKeyVVGOIdentity, &identity)))
			return
		} else {
			logger.WithField("roles", identity.Roles).WithField("path", r.URL.Path).Info("http server: access denied")
			http_helpers.WriteAPIResponse(ctx, w, http_helpers.NewUnauthorizedError())
		}
	})
}
