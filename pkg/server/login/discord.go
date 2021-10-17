package login

import (
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"io"
	"net/http"
	"os"
)

// Discord
// If the discord identity is a member of the vvgo discord server and has the vvgo-member role,
// authentication is established and a login session cookie is sent in the response.
// Otherwise, 401 unauthorized.
func Discord(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.FormValue("state") == "" {
		state, ok := oauthRedirect(w, r)
		if !ok {
			http_helpers.InternalServerError(ctx, w)
			return
		}
		http.Redirect(w, r, discord.LoginURL(state), http.StatusFound)
		return
	}

	handleError := func(err error) bool {
		if err != nil {
			logger.WithError(err).Error("discord authentication failed")
			fileName := "public/discord_login_failure.html"
			file, err := os.Open(fileName)
			w.WriteHeader(http.StatusUnauthorized)
			if err != nil {
				logger.WithField("file_name", fileName).OpenFileFailure(ctx, err)
				http_helpers.Unauthorized(ctx, w)
			}
			defer file.Close()
			_, _ = io.Copy(w, file)
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
