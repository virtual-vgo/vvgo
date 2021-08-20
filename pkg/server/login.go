package server

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/views"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strconv"
	"time"
)

const LoginCookieDuration = 2 * 7 * 24 * 3600 * time.Second // 2 weeks

type LoginRedirect struct{}

func (LoginRedirect) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	redirect := "/parts"
	if cookie, err := r.Cookie(views.CookieLoginRedirect); err != nil {
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

func loginSuccess(w http.ResponseWriter, r *http.Request, identity *login.Identity) {
	ctx := r.Context()
	cookie, err := login.NewCookie(ctx, identity, LoginCookieDuration)
	if err != nil {
		logger.WithError(err).Error("store.NewCookie() failed")
		helpers.InternalServerError(w)
		return
	}

	http.SetCookie(w, cookie)
	logger.WithFields(logrus.Fields{
		"identity": identity.Kind,
		"roles":    identity.Roles,
	}).Info("authorization succeeded")

	views.LoginSuccessView{}.ServeHTTP(w, r)
}

// PasswordLoginHandler authenticates requests using form values user and pass and a static map of valid combinations.
// If the user pass combo exists in the map, then a login cookie with the mapped roles is sent in the response.
type PasswordLoginHandler struct{}

func (x PasswordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		helpers.MethodNotAllowed(w)
		return
	}

	passwords := make(map[string]string)
	passwords["vvgo-member"] = parse_config.Config.VVGO.MemberPasswordHash

	var identity login.Identity
	if err := login.ReadSessionFromRequest(ctx, r, &identity); err == nil {
		http.Redirect(w, r, "/parts", http.StatusFound)
		return
	}

	user := r.FormValue("user")
	pass := r.FormValue("pass")
	var err error
	switch {
	case user == "":
		err = errors.New("user is required")
	case pass == "":
		err = errors.New("password is required")
	case passwords[user] == "":
		err = errors.New("unknown user")
	default:
		err = bcrypt.CompareHashAndPassword([]byte(passwords[user]), []byte(pass))
	}

	if err != nil {
		logger.WithError(err).WithField("user", user).Error("password authentication failed")
		helpers.Unauthorized(w)
		return
	}

	loginSuccess(w, r.WithContext(ctx), &login.Identity{
		Kind:  login.KindPassword,
		Roles: []login.Role{login.RoleVVGOMember},
	})
}

// DiscordLoginHandler
// If the discord identity is a member of the vvgo discord server and has the vvgo-member role,
// authentication is established and a login session cookie is sent in the response.
// Otherwise, 401 unauthorized.
type DiscordLoginHandler struct{}

var ErrNotAMember = errors.New("not a member")

func (x DiscordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.FormValue("state") == "" {
		state, ok := oauthRedirect(w, r)
		if !ok {
			helpers.InternalServerError(w)
			return
		}
		http.Redirect(w, r, discord.LoginURL(state), http.StatusFound)
		return
	}

	handleError := func(err error) bool {
		if err != nil {
			logger.WithError(err).Error("discord authentication failed")
			helpers.Unauthorized(w)
			return false
		}
		return true
	}

	if ok := handleError(validateState(r, ctx)); !ok {
		return
	}

	// get an oauth token from discord
	code := r.FormValue("code")
	oauthToken, err := discord.QueryOAuth(ctx, code)
	if ok := handleError(err); !ok {
		return
	}

	// get the user id
	discordUser, err := discord.QueryIdentity(ctx, oauthToken)
	if ok := handleError(err); !ok {
		return
	}

	// check if this user is in our guild
	guildMember, err := discord.QueryGuildMember(ctx, discordUser.ID)
	if ok := handleError(err); !ok {
		return
	}

	// check that they have the member role
	var loginRoles []login.Role
	for _, discordRole := range guildMember.Roles {
		switch discordRole {
		case "": // ignore empty strings
			continue
		case discord.VVGOExecutiveDirectorRoleID:
			loginRoles = append(loginRoles, login.RoleVVGOLeader)
		case discord.VVGOProductionTeamRoleID:
			loginRoles = append(loginRoles, login.RoleVVGOTeams)
		case discord.VVGOVerifiedMemberRoleID:
			loginRoles = append(loginRoles, login.RoleVVGOMember)
		}
	}
	if len(loginRoles) == 0 {
		handleError(ErrNotAMember)
		return
	}

	loginSuccess(w, r, &login.Identity{
		Kind:      login.KindDiscord,
		Roles:     loginRoles,
		DiscordID: discordUser.ID.String(),
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

	if err := login.DeleteSessionFromRequest(ctx, r); err != nil {
		logger.WithError(err).Error("x.Sessions.DeleteSessionFromRequest failed")
		helpers.InternalServerError(w)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
