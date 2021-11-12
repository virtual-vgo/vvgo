package guild_members

import (
	"context"
	"encoding/json"
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
)

const RedisKey = "guild_members"

type LookupRequest []string

func HandleLookup(r *http.Request) http2.Response {
	ctx := r.Context()
	var ids LookupRequest

	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		return errors.NewJsonDecodeError(err)
	} else if len(ids) == 0 {
		return errors.NewBadRequestError("ids required")
	}

	members := make([]discord.GuildMember, 0, len(ids))
	for _, id := range ids {
		var memberJSON string
		if err := redis.Do(ctx, redis.Cmd(&memberJSON, "HGET", "guild_members", id)); err != nil {
			return errors.NewInternalServerError()
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
	return http2.Response{Status: http2.StatusOk, GuildMembers: members}
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
