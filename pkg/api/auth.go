package api

import (
	"encoding/base64"
	"net/http"
	"strings"
)

type basicAuth map[string]string

func (x basicAuth) Authenticate(handlerFunc HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ok := func() bool {
			if len(x) == 0 { // skip auth for empty map
				return true
			}
			auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(auth) != 2 || auth[0] != "Basic" {
				return false
			}
			payload, _ := base64.StdEncoding.DecodeString(auth[1])
			creds := strings.SplitN(string(payload), ":", 2)
			return len(creds) == 2 && x[creds[0]] == creds[1]
		}(); !ok {
			w.Header().Set("WWW-Authenticate", `Basic charset="UTF-8"`)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
		} else {
			handlerFunc(w, r)
		}
	}
}
