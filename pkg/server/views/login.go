package views

import (
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
	"time"
)

type LoginView struct{}

const CookieLoginRedirect = "vvgo-login-redirect"

func (x LoginView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if target := r.FormValue("target"); target != "" {
		value := login.NewCookieValue()
		if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", "vvgo_login_redirect"+":"+value, "3600", target)); err != nil {
			logger.RedisFailure(ctx, err)
		} else {
			http.SetCookie(w, &http.Cookie{
				Name:     CookieLoginRedirect,
				Value:    value,
				Expires:  time.Now().Add(3600 * time.Second),
				Domain:   login.CookieDomain(),
				SameSite: http.SameSiteStrictMode,
				HttpOnly: true,
			})
		}
	}

	identity := login.IdentityFromContext(ctx)
	if identity.IsAnonymous() == false {
		http.Redirect(w, r, "/login/success", http.StatusFound)
		return
	}
	ParseAndExecute(ctx, w, r, nil, "login.gohtml")
}

type LoginSuccessView struct{}

func (x LoginSuccessView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ParseAndExecute(r.Context(), w, r, nil, "login_success.gohtml")
}
