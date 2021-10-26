package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
)

type ResolveUsersRequest []string

func ResolveUsers(r *http.Request) models.ApiResponse {
	var ids []string
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		return http_helpers.NewJsonDecodeError(err)
	}

	ctx := r.Context()
	var results []discord.GuildMember
	for _, id := range ids {
		var userJSON string
		if err := redis.Do(ctx, redis.Cmd(&userJSON, "GET", "guild_members", id)); err != nil {

		}

		user, err := discord.GetGuildMember(ctx, discord.Snowflake(id))
		if err != nil {
			logger.MethodFailure(ctx, "discord.GetGuildMember", err)
			continue
		}
		results = append(results, *user)
	}
	return models.ApiResponse{
		Status:       models.StatusOk,
		GuildMembers: results,
	}
}
