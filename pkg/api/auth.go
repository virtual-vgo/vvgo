package api

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

const HeaderVirtualVGOApiToken = "Virtual-VGO-Api-Token"

type AuthServer struct{}

type TokenAuth []Token

type Token [4]uint64

var ErrInvalidToken = errors.New("invalid token")

func NewToken() Token {
	var token Token
	for i := range token {
		binary.Read(rand.Reader, binary.LittleEndian, &token[i])
	}
	return token
}

func (token Token) String() string {
	var got [len(token)]string
	for i := range token {
		got[i] = fmt.Sprintf("%020d", token[i])
	}
	return strings.Join(got[:], "-")
}

func DecodeToken(tokenString string) (Token, error) {

	var got []string
	for i := range token {
		got[i] = fmt.Sprintf("%020d", token[i])
	}
	return strings.Join(got[:], "-")

	var token Token
	if _, err := fmt.Sscanf(tokenString, "%020d-%020d-%020d-%020d", &token[0], &token[1], &token[2], &token[3]); err != nil {
		return Token{}, ErrInvalidToken
	}
	return token, nil
}

func (token Token) Validate() error {
	switch uint64(0) {
	case token[0], token[1], token[2], token[3]:
		return ErrInvalidToken
	default:
		return nil
	}
}

func (tokens TokenAuth) AuthenticateToken(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(tokens) == 0 { // skip auth for empty slice
			handler.ServeHTTP(w, r)
			return
		}

		requestToken := r.Header.Get("VVGO-Api-Token")
		for _, token := range tokens {
			if requestToken == token.String() {
				handler.ServeHTTP(w, r)
				return
			}
		}

		logger.WithFields(logrus.Fields{
			"header": HeaderVirtualVGOApiToken,
		}).Error("token authentication failed")
		unauthorized(w)
	})
}

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
			if x[user] != pass {
				logger.WithFields(logrus.Fields{
					"user": user,
				}).Error("user authentication failed")
			}
			return x[user] == pass
		}(); !ok {
			w.Header().Set("WWW-Authenticate", `Basic charset="UTF-8"`)
			unauthorized(w)
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}
