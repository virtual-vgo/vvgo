package api

import (
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"strings"
)

const HeaderVirtualVGOApiToken = "Virtual-VGO-Api-Token"

type PassThrough struct{}

func (x PassThrough) Authenticate(handler http.Handler) http.Handler {
	return handler
}

// Authenticates http requests using basic auth.
// User name is the map key, and password is the value.
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
				logger.WithFields(logrus.Fields{
					"user": user,
				}).Error("user authentication failed")
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
