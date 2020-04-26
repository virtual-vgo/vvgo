package api

import (
	"github.com/virtual-vgo/vvgo/pkg/access"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
)

// Authenticate http requests using the sessions api
// If the request has a valid session and the required role, it is allowed access.
type RBACMux struct {
	Sessions *access.Store
	*http.ServeMux
}

func NewRBACMux(store *access.Store) *RBACMux {
	return &RBACMux{
		Sessions: store,
		ServeMux: http.NewServeMux(),
	}
}

func (auth *RBACMux) Handle(pattern string, handler http.Handler, role access.Role) {
	// anonymous access goes directly to the mux
	if role == access.RoleAnonymous {
		auth.ServeMux.Handle(pattern, handler)
		return
	}

	auth.ServeMux.Handle(pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorized := func() bool {
			ctx, span := tracing.StartSpan(r.Context(), "rbac_mux")
			defer span.Send()

			var identity access.Identity
			if err := auth.Sessions.ReadIdentityFromRequest(ctx, r, &identity); err != nil {
				return false
			}

			for _, gotRole := range identity.Roles {
				if role == gotRole {
					return true
				}
			}
			return false
		}()

		switch {
		case authorized:
			handler.ServeHTTP(w, r)
		case acceptsType(r, "text/html"):
			http.Redirect(w, r, "/login", http.StatusFound)
		default:
			unauthorized(w)
		}
	}))
}
