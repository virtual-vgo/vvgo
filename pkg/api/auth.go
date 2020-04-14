package api

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

const HeaderVirtualVGOApiToken = "Virtual-VGO-Api-Token"

type AuthServer struct{}

type TokenAuth []Token

func (tokens TokenAuth) Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok := func() bool {
			if len(tokens) == 0 { // skip auth for empty slice
				return true
			}
			requestToken := r.Header.Get(HeaderVirtualVGOApiToken)
			for _, token := range tokens {
				if requestToken == token.String() {
					return true
				}
			}
			return false
		}(); ok {
			handler.ServeHTTP(w, r)
		} else {
			logger.WithFields(logrus.Fields{
				"header": HeaderVirtualVGOApiToken,
			}).Error("token authentication failed")
			unauthorized(w)
		}
	})
}

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
		got[i] = fmt.Sprintf("%016x", token[i])
	}
	return strings.Join(got[:], "-")
}

func DecodeToken(tokenString string) (Token, error) {
	tokenParts := strings.Split(tokenString, "-")
	var token Token
	if len(tokenParts) != len(token) {
		return Token{}, ErrInvalidToken
	}
	for i := range token {
		if len(tokenParts[i]) != 16 {
			return Token{}, ErrInvalidToken
		}
		token[i], _ = strconv.ParseUint(tokenParts[i], 16, 64)
	}
	return token, token.Validate()
}

func (token Token) Validate() error {
	for i := range token {
		if token[i] == 0 {
			return ErrInvalidToken
		}
	}
	return nil
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
