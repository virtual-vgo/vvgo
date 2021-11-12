package guild_members

import (
	"bytes"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/cache"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"time"
)

var HandleList = cache.Handle(4*3600*time.Second, func(r *http.Request) api.Response {
	ctx := r.Context()
	members, err := discord.ListGuildMembers(ctx, 1000, 0)
	if err != nil {
		if e, ok := err.(*discord.Error); ok {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(e); err != nil {
				logger.JsonEncodeFailure(ctx, err)
			}
			return response.NewErrorResponse(response.Error{
				Code:    e.Code,
				Message: e.Error(),
				Data:    buf.Bytes(),
			})
		}
		return response.NewInternalServerError()
	}

	verified := make([]discord.GuildMember, 0, len(members))
	for _, member := range members {
		for _, role := range member.Roles {
			if role == auth.RoleVVGOVerifiedMember.String() {
				verified = append(verified, member)
				break
			}
		}
	}

	return api.Response{Status: api.StatusOk, GuildMembers: members}
})
