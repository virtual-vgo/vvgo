package api

import (
	"errors"
	"github.com/virtual-vgo/vvgo/pkg/access"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
)

type DiscordAuthHandler struct {
	GuildID        discord.GuildID
	RoleVVGOMember string
	Sessions       *sessions.Store
	Discord        *discord.Client
}

var ErrNotAMember = errors.New("not a member")

func (x DiscordAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	identity := sessions.Identity{
		Kind:        sessions.KindDiscord,
		Roles:       []access.Role{access.RoleVVGOMember},
	}
	loginRedirect(ctx, w, r, x.Sessions, &identity)
}
