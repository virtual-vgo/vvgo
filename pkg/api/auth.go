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

// Authenticates http requests using basic auth.
// User name is the map key, and password is the value.
// If the map is empty or nil, requests are always authenticated.
type BasicAuth map[string]string

func (auth BasicAuth) Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok := func() bool {
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
		}(); !ok {
			w.Header().Set("WWW-Authenticate", `Basic charset="UTF-8"`)
			unauthorized(w)
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}

type TokenAuth []string

func (tokens TokenAuth) Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok := func() bool {
			requestToken := r.Header.Get(HeaderVirtualVGOApiToken)
			for _, token := range tokens {
				if requestToken == token {
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
