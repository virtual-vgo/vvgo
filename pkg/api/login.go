package api

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/access"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
)

type PasswordLoginHandler struct {
	Sessions *access.Store
	Logins   []PasswordLogin
}

type PasswordLogin struct {
	User  string
	Pass  string
	Roles []access.Role
}

func (x PasswordLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "password_login")
	defer span.Send()

	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var identity access.Identity
	if err := x.Sessions.ReadIdentityFromRequest(ctx, r, &identity); err == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	user := r.FormValue("user")
	pass := r.FormValue("pass")

	var roles []access.Role
	for _, login := range x.Logins {
		if user == login.User && pass == login.Pass {
			roles = login.Roles
		}
	}

	if len(roles) == 0 {
		logger.WithFields(logrus.Fields{
			"user": user,
		}).Error("password authentication failed")
		unauthorized(w)
		return
	}

	// create the identity object
	identity = access.Identity{
		Kind:  access.KindPassword,
		Roles: roles,
	}
	loginRedirect(ctx, w, r, x.Sessions, &identity)
}

type DiscordLoginHandler struct {
	GuildID        discord.GuildID
	RoleVVGOMember string
	Sessions       *access.Store
	Discord        *discord.Client
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

	// get an oauth token from discord
	code := r.FormValue("code")
	oauthToken, err := x.Discord.QueryOAuth(ctx, code)
	if ok := handleError(err); !ok {
		return
	}

	// get the user id
	discordUser, err := x.Discord.QueryIdentity(ctx, oauthToken)
	if ok := handleError(err); !ok {
		return
	}

	// check if this user is in our guild
	guildMember, err := x.Discord.QueryGuildMember(ctx, x.GuildID, discordUser.ID)
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

	// create the identity object
	identity := access.Identity{
		Kind:  access.KindDiscord,
		Roles: []access.Role{access.RoleVVGOMember},
	}
	loginRedirect(ctx, w, r, x.Sessions, &identity)
}

type LogoutHandler struct {
	Sessions *access.Store
}

func (x LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "logout_handler")
	defer span.Send()

	var session access.Session
	if err := x.Sessions.ReadSessionFromRequest(r, &session); err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// delete the session
	if err := x.Sessions.DeleteIdentity(ctx, session.ID); err != nil {
		logger.WithError(err).Error("x.Sessions.DeleteIdentity() failed")
		internalServerError(w)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
	return
}
