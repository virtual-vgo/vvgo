package login

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"strconv"
	"time"
)

var ErrNotAMember = errors.New("not a member")

const (
	SessionCookieName     = "vvgo-sessions"
	SessionCookiePath     = "/"
	SessionCookieDuration = 2 * 7 * 24 * 3600 * time.Second // 2 weeks
	RedirectCookieName    = "vvgo-login-redirect"
	CtxKeyVVGOIdentity    = "vvgo_identity"
)

func Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	redirect := "/parts"
	if cookie, err := r.Cookie(RedirectCookieName); err != nil {
		logger.MethodFailure(ctx, "r.Cookie", err)
	} else {
		var want string
		if err := redis.Do(ctx, redis.Cmd(&want, "GET", "vvgo_login_redirect"+":"+cookie.Value)); err != nil {
			logger.RedisFailure(ctx, err)
		} else {
			redirect = want
		}
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func loginSuccess(w http.ResponseWriter, r *http.Request, identity *models.Identity) {
	ctx := r.Context()
	cookie, err := NewCookie(ctx, identity, SessionCookieDuration)
	if err != nil {
		logger.NewCookieFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
		return
	}

	http.SetCookie(w, cookie)
	logger.WithFields(logrus.Fields{
		"identity": identity.Kind,
		"roles":    identity.Roles,
	}).Info("authorization succeeded")

	http.Redirect(w, r, "/login/success", http.StatusFound)
}

const CookieOAuthState = "vvgo-oauth-state"

func oauthRedirect(w http.ResponseWriter, r *http.Request) (string, bool) {
	ctx := r.Context()

	// read a random state number
	statusBytes := make([]byte, 32)
	if _, err := rand.Read(statusBytes); err != nil {
		logger.MethodFailure(ctx, "rand.Read", err)
		return "", false
	}
	state := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[:16]), 16)
	value := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[16:]), 16)

	// store the number in redis
	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", "oauth_state:"+state, "300", value)); err != nil {
		logger.RedisFailure(ctx, err)
		return "", false
	}
	http.SetCookie(w, &http.Cookie{
		Name:    CookieOAuthState,
		Value:   value,
		Expires: time.Now().Add(300 * time.Second),
	})
	return state, true
}

func validateState(r *http.Request, ctx context.Context) error {
	state := r.FormValue("state")
	if state == "" {
		return errors.New("no state param")
	}

	// check if it exists in redis
	var value string
	if err := redis.Do(ctx, redis.Cmd(&value, "GET", "oauth_state:"+state)); err != nil {
		return err
	}

	// check against the cookie value
	cookie, err := r.Cookie(CookieOAuthState)
	if err != nil {
		return err
	}
	if cookie.Value != value {
		return errors.New("invalid state")
	}
	return nil
}

// Logout deletes the login session from the incoming request, if it exists.
func Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := DeleteSessionFromRequest(ctx, r); err != nil {
		logger.MethodFailure(ctx, "login.DeleteSessionFromRequest", err)
		http_helpers.InternalServerError(ctx, w)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
