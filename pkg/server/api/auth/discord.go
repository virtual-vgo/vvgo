package auth

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

// Discord
// If the discord identity is a member of the vvgo discord server and has the vvgo-member role,
// authentication is established and a login session cookie is sent in the response.
// Otherwise, 401 unauthorized.

type PostDiscordRequest struct {
	Code   string `json:"code"`
	State  string `json:"state"`
	Secret string `json:"secret"`
}

func Discord(r *http.Request) models.ApiResponse {
	ctx := r.Context()

	logAuthFailure := func(reason string) {
		logger.WithField("reason", reason).Error("discord authentication failed")
	}

	var data PostDiscordRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	switch {
	case err != nil:
		logAuthFailure("json decode error")
		return http_helpers.NewJsonDecodeError(err)
	case data.State == "":
		logAuthFailure("state is required")
		return http_helpers.NewBadRequestError("state is required")
	case data.Code == "":
		logAuthFailure("code is required")
		return http_helpers.NewBadRequestError("code is required")
	case data.Secret == "":
		logAuthFailure("secret is required")
		return http_helpers.NewBadRequestError("secret is required")
	}

	if !validateState(ctx, data.State, data.Secret) {
		logAuthFailure("invalid state")
		return http_helpers.NewUnauthorizedError()
	}

	// get an oauth token from discord
	oauthToken, err := discord.GetOAuthToken(ctx, data.Code)
	if err != nil {
		logger.MethodFailure(ctx, "discord.GetOAuthToken", err)
		logAuthFailure("internal server error")
		return http_helpers.NewUnauthorizedError()
	}

	// get the user id
	discordUser, err := discord.GetIdentity(ctx, oauthToken)
	if err != nil {
		logger.MethodFailure(ctx, "discord.GetIdentity", err)
		logAuthFailure("internal server error")
		return http_helpers.NewUnauthorizedError()
	}

	// check if this user is in our guild
	guildMember, err := discord.GetGuildMember(ctx, discordUser.ID)
	if err != nil {
		logger.MethodFailure(ctx, "discord.GetGuildMember", err)
		logAuthFailure("not a member")
		return http_helpers.NewUnauthorizedError()
	}

	// check that they have the member role
	var loginRoles []models.Role
	for _, discordRole := range guildMember.Roles {
		switch discordRole {
		case "": // ignore empty strings
			continue
		case discord.VVGOExecutiveDirectorRoleID:
			loginRoles = append(loginRoles, models.RoleVVGOExecutiveDirector)
		case discord.VVGOProductionTeamRoleID:
			loginRoles = append(loginRoles, models.RoleVVGOProductionTeam)
		case discord.VVGOVerifiedMemberRoleID:
			loginRoles = append(loginRoles, models.RoleVVGOVerifiedMember)
		}
	}

	if len(loginRoles) == 0 {
		logAuthFailure("not a member")
		return http_helpers.NewUnauthorizedError()
	}

	identity := models.Identity{
		Kind:      models.KindDiscord,
		Roles:     loginRoles,
		DiscordID: discordUser.ID.String(),
	}

	if _, err := login.NewSession(ctx, &identity, SessionDuration); err != nil {
		logger.MethodFailure(ctx, "login.NewSession", err)
		logAuthFailure("internal server error")
		return http_helpers.NewInternalServerError()
	}

	return models.ApiResponse{Status: models.StatusOk, Identity: &identity}
}
