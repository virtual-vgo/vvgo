package api

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
	"strconv"
	"time"
)

const LoginCookieDuration = 2 * 7 * 24 * 3600 * time.Second // 2 weeks

type LoginView struct{}

const CookieLoginRedirect = "vvgo-login-redirect"

func (x LoginView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if target := r.FormValue("target"); target != "" {
		value := login.NewCookieValue()
		if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", "vvgo_login_redirect"+":"+value, "3600", target)); err != nil {
			logger.WithError(err).Error("redis.Do() failed")
		} else {
			http.SetCookie(w, &http.Cookie{
				Name:     CookieLoginRedirect,
				Value:    value,
				Expires:  time.Now().Add(3600 * time.Second),
				Domain:   login.NewStore(ctx).Config().CookieDomain,
				SameSite: http.SameSiteStrictMode,
				HttpOnly: true,
			})
		}
	}

	identity := IdentityFromContext(ctx)
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

type LoginRedirect struct{}

func (LoginRedirect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	redirect := "/parts"
	if cookie, err := r.Cookie(CookieLoginRedirect); err != nil {
		logger.WithError(err).Error("r.Cookie() failed")
	} else {
		var want string
		if err := redis.Do(ctx, redis.Cmd(&want, "GET", "vvgo_login_redirect"+":"+cookie.Value)); err != nil {
			logger.WithError(err).Error("redis.Do() failed")
		} else {
			redirect = want
		}
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func loginSuccess(w http.ResponseWriter, r *http.Request, ctx context.Context, identity *login.Identity) {
	cookie, err := login.NewStore(ctx).NewCookie(ctx, identity, LoginCookieDuration)
	if err != nil {
		logger.WithError(err).Error("store.NewCookie() failed")
		internalServerError(w)
		return
	}

	http.SetCookie(w, cookie)
	logger.WithFields(logrus.Fields{
		"identity": identity.Kind,
		"roles":    identity.Roles,
	}).Info("authorization succeeded")

	LoginSuccessView{}.ServeHTTP(w, r)
}

// PasswordLoginHandler authenticates requests using form values user and pass and a static map of valid combinations.
// If the user pass combo exists in the map, then a login cookie with the mapped roles is create and sent in the response.
type PasswordLoginHandler struct{}

func (x PasswordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	passwords := make(map[string]string)
	if err := parse_config.ReadFromRedisHash(ctx, "password_login", &passwords); err != nil {
		logger.WithError(err).Errorf("redis.Do() failed: %v", err)
		internalServerError(w)
		return
	}

	var identity login.Identity
	if err := login.NewStore(ctx).ReadSessionFromRequest(ctx, r, &identity); err == nil {
		http.Redirect(w, r, "/parts", http.StatusFound)
		return
	}

	user := r.FormValue("user")
	pass := r.FormValue("pass")

	if user == "" || pass == "" || passwords[user] != pass {
		logger.WithFields(logrus.Fields{
			"user": user,
		}).Error("password authentication failed")
		unauthorized(w)
		return
	}

	loginSuccess(w, r, ctx, &login.Identity{
		Kind:  login.KindPassword,
		Roles: []login.Role{login.RoleVVGOMember},
	})
}

// DiscordLoginHandler
// If the discord identity is a member of the vvgo discord server and has the vvgo-member role,
// authentication is established and a login session cookie is sent in the response.
// Otherwise, 401 unauthorized.
type DiscordLoginHandler struct{}

type DiscordLoginConfig struct {
	GuildID          string `redis:"guild_id"`
	RoleVVGOMemberID string `redis:"role_vvgo_member"`
	RoleVVGOTeamsID  string `redis:"role_vvgo_teams"`
	RoleVVGOLeaderID string `redis:"role_vvgo_leader"`
}

var ErrNotAMember = errors.New("not a member")

func (x DiscordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.FormValue("state") == "" {
		state, ok := oauthRedirect(w, r)
		if !ok {
			internalServerError(w)
			return
		}
		http.Redirect(w, r, discord.NewClient(ctx).LoginURL(state), http.StatusFound)
	} else {
		var config DiscordLoginConfig
		if err := parse_config.ReadFromRedisHash(ctx, "discord_login", &config); err != nil {
			logger.WithError(err).Errorf("redis.Do() failed: %v", err)
			internalServerError(w)
		}
		x.authorize(w, r, config)
	}
}

func (x DiscordLoginHandler) authorize(w http.ResponseWriter, r *http.Request, config DiscordLoginConfig) {
	ctx := r.Context()
	discordClient := discord.NewClient(ctx)

	handleError := func(err error) bool {
		if err != nil {
			logger.WithError(err).Error("discord authentication failed")
			unauthorized(w)
			return false
		}
		return true
	}

	if ok := handleError(validateState(r, ctx)); !ok {
		return
	}

	// get an oauth token from discord
	code := r.FormValue("code")
	oauthToken, err := discordClient.QueryOAuth(ctx, code)
	if ok := handleError(err); !ok {
		return
	}

	// get the user id
	discordUser, err := discordClient.QueryIdentity(ctx, oauthToken)
	if ok := handleError(err); !ok {
		return
	}

	// check if this user is in our guild
	guildMember, err := discordClient.QueryGuildMember(ctx, discord.GuildID(config.GuildID), discordUser.ID)
	if ok := handleError(err); !ok {
		return
	}

	// check that they have the member role
	var loginRoles []login.Role
	for _, discordRole := range guildMember.Roles {
		switch discordRole {
		case "": // ignore empty strings
			continue
		case config.RoleVVGOLeaderID:
			loginRoles = append(loginRoles, login.RoleVVGOLeader)
		case config.RoleVVGOTeamsID:
			loginRoles = append(loginRoles, login.RoleVVGOTeams)
		case config.RoleVVGOMemberID:
			loginRoles = append(loginRoles, login.RoleVVGOMember)
		}
	}
	if len(loginRoles) == 0 {
		handleError(ErrNotAMember)
		return
	}

	loginSuccess(w, r, ctx, &login.Identity{
		Kind:  login.KindDiscord,
		Roles: loginRoles,
	})
}

const CookieOAuthState = "vvgo-oauth-state"

func oauthRedirect(w http.ResponseWriter, r *http.Request) (string, bool) {
	ctx := r.Context()

	// read a random state number
	statusBytes := make([]byte, 32)
	if _, err := rand.Read(statusBytes); err != nil {
		logger.WithError(err).Error("rand.Read() failed")
		return "", false
	}
	state := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[:16]), 16)
	value := strconv.FormatUint(binary.BigEndian.Uint64(statusBytes[16:]), 16)

	// store the number in redis
	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", "oauth_state:"+state, "300", value)); err != nil {
		logger.WithError(err).Error("redis.Do() failed")
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

// LogoutHandler deletes the login session from the incoming request, if it exists.
type LogoutHandler struct{}

func (x LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := login.NewStore(ctx).DeleteSessionFromRequest(ctx, r); err != nil {
		logger.WithError(err).Error("x.Sessions.DeleteSessionFromRequest failed")
		internalServerError(w)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
