package guild_members

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
)

const RedisKey = "guild_members"

type LookupRequest []string

func HandleLookup(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	var ids LookupRequest

	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		return http_helpers.NewJsonDecodeError(err)
	} else if len(ids) == 0 {
		return http_helpers.NewBadRequestError("ids required")
	}

	members := make([]discord.GuildMember, 0, len(ids))
	for _, id := range ids {
		var memberJSON string
		if err := redis.Do(ctx, redis.Cmd(&memberJSON, "HGET", "guild_members", id)); err != nil {
			return http_helpers.NewInternalServerError()
		}

		var guildMember *discord.GuildMember
		if memberJSON != "" {
			if err := json.Unmarshal([]byte(memberJSON), guildMember); err != nil {
				logger.JsonDecodeFailure(ctx, err)
				guildMember = nil
			}
		}

		if guildMember == nil {
			var err error
			guildMember, err = discord.GetGuildMember(ctx, discord.Snowflake(id))
			if err != nil {
				logger.MethodFailure(ctx, "discord.GetGuildMember", err)
				guildMember = nil
			}
		}

		if guildMember != nil {
			members = append(members, *guildMember)
		}
	}

	saveGuildMembers(ctx, members)
	return models.ApiResponse{Status: models.StatusOk, GuildMembers: members}
}

func saveGuildMembers(ctx context.Context, members []discord.GuildMember) {
	args := []string{RedisKey}
	for _, member := range members {
		memberJSON, err := json.Marshal(member)
		if err != nil {
			logger.JsonEncodeFailure(ctx, err)
			continue
		}
		args = append(args, member.User.ID.String(), string(memberJSON))
	}
	if err := redis.Do(ctx, redis.Cmd(nil, "HSET", args...)); err != nil {
		logger.RedisFailure(ctx, err)
	}
}
