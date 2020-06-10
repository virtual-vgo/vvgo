package api

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"time"
)

const LoginCookieDuration = 2 * 7 * 24 * 3600 * time.Second // 2 weeks

// PasswordLoginHandler authenticates requests using form values user and pass and a static map of valid combinations.
// If the user pass combo exists in the map, then a login cookie with the mapped roles is create and sent in the response.
type PasswordLoginHandler struct {
	Sessions *login.Store

	// Logins is a map of login user and pass to a slice of roles for that login.
	Logins map[[2]string][]login.Role
}

func (x PasswordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "password_login")
	defer span.Send()

	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var identity login.Identity
	if err := x.Sessions.ReadSessionFromRequest(ctx, r, &identity); err == nil {
		http.Redirect(w, r, "/", http.StatusFound)
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

	identity = login.Identity{
		Kind:  login.KindPassword,
		Roles: gotRoles,
	}

	cookie, err := x.Sessions.NewCookie(ctx, &identity, LoginCookieDuration)
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
	http.Redirect(w, r, "/", http.StatusFound)
}

// DiscordLoginHandler accepts an oauth token in the request body and uses the token to query for discord identity.
// If the discord identity is a member of the vvgo discord server and has the vvgo-member role,
// authentication is established and a login session cookie is sent in the response.
// Otherwise, 401 unauthorized.
type DiscordLoginHandler struct {
	GuildID        discord.GuildID
	RoleVVGOMember string
	Sessions       *login.Store
}

var ErrNotAMember = errors.New("not a member")

func (x DiscordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "discord_oauth_handler")
	defer span.Send()

	handleError := func(err error) bool {
		if err != nil {
			tracing.AddError(ctx, err)
			logger.WithError(err).Error("discord authentication failed")
			unauthorized(w)
			return false
		}
		return true
	}

	// read the oauth token from the request
	var oauthToken discord.OAuthToken
	if ok := handleError(json.NewDecoder(r.Body).Decode(&oauthToken)); !ok {
		return
	}

	// get the user id
	discordUser, err := discord.QueryIdentity(ctx, &oauthToken)
	if ok := handleError(err); !ok {
		return
	}

	// check if this user is in our guild
	guildMember, err := discord.QueryGuildMember(ctx, x.GuildID, discordUser.ID)
	if ok := handleError(err); !ok {
		return
	}

	// check that they have the member role
	var ok bool
	for _, role := range guildMember.Roles {
		if role == x.RoleVVGOMember {
			ok = true
			break
		}
	}
	if !ok {
		handleError(ErrNotAMember)
		return
	}

	// Create a login session
	cookie, err := x.Sessions.NewCookie(ctx, &login.Identity{
		Kind:  login.KindDiscord,
		Roles: []login.Role{login.RoleVVGOMember},
	}, LoginCookieDuration)
	if err != nil {
		logger.WithError(err).Error("sessions.NewCookie() failed")
		internalServerError(w)
		return
	}

	// redirect to home
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

// LogoutHandler deletes the login session from the incoming request, if it exists.
type LogoutHandler struct {
	Sessions *login.Store
}

func (x LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "logout_handler")
	defer span.Send()

	if err := x.Sessions.DeleteSessionFromRequest(ctx, r); err != nil {
		logger.WithError(err).Error("x.Sessions.DeleteSessionFromRequest failed")
		internalServerError(w)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
