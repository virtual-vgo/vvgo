package login

import (
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"net/http"
)

// DiscordLoginHandler
// If the discord identity is a member of the vvgo discord server and has the vvgo-member role,
// authentication is established and a login session cookie is sent in the response.
// Otherwise, 401 unauthorized.
func DiscordLoginHandler(w http.ResponseWriter, r *http.Request) {
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
	var loginRoles []models.Role
	for _, discordRole := range guildMember.Roles {
		switch discordRole {
		case "": // ignore empty strings
			continue
		case discord.VVGOExecutiveDirectorRoleID:
			loginRoles = append(loginRoles, models.RoleVVGOLeader)
		case discord.VVGOProductionTeamRoleID:
			loginRoles = append(loginRoles, models.RoleVVGOTeams)
		case discord.VVGOVerifiedMemberRoleID:
			loginRoles = append(loginRoles, models.RoleVVGOMember)
		}
	}
	if len(loginRoles) == 0 {
		handleError(ErrNotAMember)
		return
	}

	loginSuccess(w, r, &models.Identity{
		Kind:      models.KindDiscord,
		Roles:     loginRoles,
		DiscordID: discordUser.ID.String(),
	})
}
