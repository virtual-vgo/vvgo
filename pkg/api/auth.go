package api

import (
	"net/http"
)

type basicAuth map[string]string

func (x basicAuth) Authenticate(handlerFunc HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "authorization failed", http.StatusUnauthorized)
		} else {
			handlerFunc(w, r)
		}
	}
}
