package api

import (
	"net/http"
)

// Authenticates http requests using basic auth.
// User name is the map key, and password is the value.
// If the map is empty or nil, requests are always authenticated.
type basicAuth map[string]string

func (x basicAuth) Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok := func() bool {
			if len(x) == 0 { // skip auth for empty map
				return true
			}
			user, pass, ok := r.BasicAuth()
			if !ok || user == "" || pass == "" {
				return false
			}
			return x[user] == pass
		}(); !ok {
			w.Header().Set("WWW-Authenticate", `Basic charset="UTF-8"`)
			unauthorized(w, r)
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}
