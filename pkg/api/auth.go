package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"strings"
)

// Authenticate http requests using the sessions api
// If the request has a valid session and the required role, it is allowed access.
type RBACMux struct {
	Sessions *sessions.Store
	*http.ServeMux
}

func NewRBACMux(store *sessions.Store) *RBACMux {
	return &RBACMux{
		Sessions: store,
		ServeMux: http.NewServeMux(),
	}
}

func (auth *RBACMux) Handle(pattern string, handler http.Handler, role sessions.Role) {
	// anonymous access goes directly to the mux
	if role == sessions.RoleAnonymous {
		auth.ServeMux.Handle(pattern, handler)
		return
	}

	auth.ServeMux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok := func() bool {
			ctx, span := tracing.StartSpan(r.Context(), "rbac_mux")
			defer span.Send()

			var identity sessions.Identity
			if err := auth.Sessions.ReadIdentityFromRequest(ctx, r, &identity); err != nil {
				return false
			}

			for _, gotRole := range identity.Roles {
				if role == gotRole {
					return true
				}
			}
			return false
		}(); ok {
			handler.ServeHTTP(w, r)
		} else {
			unauthorized(w)
		}
	}))
}

type PassThrough struct{}

func (x PassThrough) Authenticate(handler http.Handler) http.Handler {
	return handler
}

// Authenticates http requests using basic auth.
// Identity name is the map key, and password is the value.
// If the map is empty or nil, requests are always authenticated.
type BasicAuth map[string]string

func (auth BasicAuth) Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ok := func() bool {
			_, span := tracing.StartSpan(ctx, "basic_auth")
			defer span.Send()
			user, pass, ok := r.BasicAuth()
			if !ok || user == "" || pass == "" {
				return false
			}
			if auth[user] == pass {
				return true
			} else {
				logger.WithField("user", user).Error("basic authentication failed")
				return false
			}
		}(); ok {
			tracing.WrapHandler(handler).ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", `Basic charset="UTF-8"`)
			unauthorized(w)
		}
	})
}

type TokenAuth []string

func (tokens TokenAuth) Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if ok := func() bool {
			_, span := tracing.StartSpan(ctx, "token_auth")
			defer span.Send()
			auth := strings.TrimSpace(r.Header.Get("Authorization"))
			for _, token := range tokens {
				if auth == "Bearer "+token {
					return true
				}
			}
			return false
		}(); ok {
			tracing.WrapHandler(handler).ServeHTTP(w, r)
		} else {
			logger.Error("token authentication failed")
			unauthorized(w)
		}
	})
}
