package api

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"net/http"
	"strconv"
	"time"
)

const LoginCookieDuration = 2 * 7 * 24 * 3600 * time.Second // 2 weeks

type LoginView struct {
	Sessions *login.Store
	Template
}

func (x LoginView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	identity := IdentityFromContext(ctx)
	if identity.IsAnonymous() == false {
		http.Redirect(w, r, "/login/success", http.StatusFound)
		return
	}
	x.Template.ParseAndExecute(ctx, w, r, nil, "login.gohtml")
}

type LoginSuccessView struct{ Template }

func (x LoginSuccessView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	x.Template.ParseAndExecute(r.Context(), w, r, nil, "login_success.gohtml")
}

func loginSuccess(w http.ResponseWriter, r *http.Request, ctx context.Context, sessions *login.Store, identity *login.Identity) {
	cookie, err := sessions.NewCookie(ctx, identity, LoginCookieDuration)
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
type PasswordLoginHandler struct {
	Sessions *login.Store

	// Logins is a map of login user and pass to a slice of roles for that login.
	Logins map[[2]string][]login.Role
}

func (x PasswordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var identity login.Identity
	if err := x.Sessions.ReadSessionFromRequest(ctx, r, &identity); err == nil {
		http.Redirect(w, r, "/parts", http.StatusFound)
		return
	}

	user := r.FormValue("user")
	pass := r.FormValue("pass")

	gotRoles, ok := x.Logins[[2]string{user, pass}]
	if !ok {
		logger.WithFields(logrus.Fields{
			"user": user,
		}).Error("password authentication failed")
		unauthorized(w)
		return
	}

	loginSuccess(w, r, ctx, x.Sessions, &login.Identity{
		Kind:  login.KindPassword,
		Roles: gotRoles,
	})
}

// DiscordLoginHandler
// If the discord identity is a member of the vvgo discord server and has the vvgo-member role,
// authentication is established and a login session cookie is sent in the response.
// Otherwise, 401 unauthorized.
type DiscordLoginHandler struct {
	GuildID          discord.GuildID
	RoleVVGOMemberID string
	RoleVVGOTeamsID  string
	RoleVVGOLeaderID string
	Sessions         *login.Store
	Namespace        string
}

var ErrNotAMember = errors.New("not a member")

func (x DiscordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") == "" {
		state, ok := oauthRedirect(w, r, x.Namespace)
		if !ok {
			internalServerError(w)
			return
		}
		http.Redirect(w, r, discord.LoginURL(state), http.StatusFound)
	} else {
		x.authorize(w, r)
	}
}

func (x DiscordLoginHandler) authorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handleError := func(err error) bool {
		if err != nil {
			logger.WithError(err).Error("discord authentication failed")
			unauthorized(w)
			return false
		}
		return true
	}

	if ok := handleError(validateState(r, ctx, x.Namespace)); !ok {
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
	guildMember, err := discord.QueryGuildMember(ctx, x.GuildID, discordUser.ID)
	if ok := handleError(err); !ok {
		return
	}

	// check that they have the member role
	var loginRoles []login.Role
	for _, discordRole := range guildMember.Roles {
		switch discordRole {
		case "": // ignore empty strings
			continue
		case x.RoleVVGOLeaderID:
			loginRoles = append(loginRoles, login.RoleVVGOLeader)
		case x.RoleVVGOTeamsID:
			loginRoles = append(loginRoles, login.RoleVVGOTeams)
		case x.RoleVVGOMemberID:
			loginRoles = append(loginRoles, login.RoleVVGOMember)
		}
	}
	if len(loginRoles) == 0 {
		handleError(ErrNotAMember)
		return
	}

	loginSuccess(w, r, ctx, x.Sessions, &login.Identity{
		Kind:  login.KindDiscord,
		Roles: loginRoles,
	})
}

const CookieOAuthState = "vvgo-oauth-state"

func oauthRedirect(w http.ResponseWriter, r *http.Request, redisNamespace string) (string, bool) {
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
	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", redisNamespace+":oauth_state:"+state, "300", value)); err != nil {
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

func validateState(r *http.Request, ctx context.Context, redisNamespace string) error {
	state := r.FormValue("state")
	if state == "" {
		return errors.New("no state param")
	}

	// check if it exists in redis
	var value string
	if err := redis.Do(ctx, redis.Cmd(&value, "GET", redisNamespace+":oauth_state:"+state)); err != nil {
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
type LogoutHandler struct {
	Sessions *login.Store
}

func (x LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := x.Sessions.DeleteSessionFromRequest(ctx, r); err != nil {
		logger.WithError(err).Error("x.Sessions.DeleteSessionFromRequest failed")
		internalServerError(w)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
