package auth

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
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

func Discord(r *http.Request) api.Response {
	ctx := r.Context()

	logAuthFailure := func(reason string) {
		logger.WithField("reason", reason).Error("discord authentication failed")
	}

	var data PostDiscordRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	switch {
	case err != nil:
		logAuthFailure("json decode error")
		return errors.NewJsonDecodeError(err)
	case data.State == "":
		logAuthFailure("state is required")
		return errors.NewBadRequestError("state is required")
	case data.Code == "":
		logAuthFailure("code is required")
		return errors.NewBadRequestError("code is required")
	case data.Secret == "":
		logAuthFailure("secret is required")
		return errors.NewBadRequestError("secret is required")
	}

	if !validateState(ctx, data.State, data.Secret) {
		logAuthFailure("invalid state")
		return errors.NewUnauthorizedError()
	}

	// get an oauth token from discord
	oauthToken, err := discord.GetOAuthToken(ctx, data.Code)
	if err != nil {
		logger.MethodFailure(ctx, "discord.GetOAuthToken", err)
		logAuthFailure("internal server error")
		return errors.NewUnauthorizedError()
	}

	// get the user id
	discordUser, err := discord.GetIdentity(ctx, oauthToken)
	if err != nil {
		logger.MethodFailure(ctx, "discord.GetIdentity", err)
		logAuthFailure("internal server error")
		return errors.NewUnauthorizedError()
	}

	// check if this user is in our guild
	guildMember, err := discord.GetGuildMember(ctx, discordUser.ID)
	if err != nil {
		logger.MethodFailure(ctx, "discord.GetGuildMember", err)
		logAuthFailure("not a member")
		return errors.NewUnauthorizedError()
	}

	// check that they have the member role
	var loginRoles []auth.Role
	for _, discordRole := range guildMember.Roles {
		switch discordRole {
		case "": // ignore empty strings
			continue
		case discord.VVGOExecutiveDirectorRoleID:
			loginRoles = append(loginRoles, auth.RoleVVGOExecutiveDirector)
		case discord.VVGOProductionTeamRoleID:
			loginRoles = append(loginRoles, auth.RoleVVGOProductionTeam)
		case discord.VVGOVerifiedMemberRoleID:
			loginRoles = append(loginRoles, auth.RoleVVGOVerifiedMember)
		}
	}

	if len(loginRoles) == 0 {
		logAuthFailure("not a member")
		return errors.NewUnauthorizedError()
	}

	identity := auth.Identity{
		Kind:      auth.KindDiscord,
		Roles:     loginRoles,
		DiscordID: discordUser.ID.String(),
	}

	if _, err := auth.NewSession(ctx, &identity, auth.SessionDuration); err != nil {
		logger.MethodFailure(ctx, "login.NewSession", err)
		logAuthFailure("internal server error")
		return errors.NewInternalServerError()
	}

	return api.Response{Status: api.StatusOk, Identity: &identity}
}
