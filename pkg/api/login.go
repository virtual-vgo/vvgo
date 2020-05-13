package api

import (
	"encoding/json"
	"errors"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"time"
)

const DiscordLoginExpires = 7 * 24 * 3600 * time.Second // 1 week

// DiscordLoginHandler accepts an oauth token in the request body and uses the token to query for discord identity.
// If the discord identity is a member of the vvgo discord server and has the vvgo-member role,
// authentication is established and a login session cookie is sent in the response.
// Otherwise, 401 unauthorized.
type DiscordLoginHandler struct {
	GuildID        discord.GuildID
	RoleVVGOMember string
	Discord        *discord.Client
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
	discordUser, err := x.Discord.QueryIdentity(ctx, &oauthToken)
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

	// Create a login session
	cookie, err := x.Sessions.NewCookie(ctx, &login.Identity{
		Kind:  login.KindDiscord,
		Roles: []login.Role{login.RoleVVGOMember},
	}, DiscordLoginExpires)
	if err != nil {
		logger.WithError(err).Error("sessions.NewCookie() failed")
		internalServerError(w)
		return
	}

	// redirect to home
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}
