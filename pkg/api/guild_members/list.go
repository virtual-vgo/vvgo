package guild_members

import (
	"bytes"
	"encoding/json"
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/cache"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"time"
)

var HandleList = cache.Handle(4*3600*time.Second, func(r *http.Request) http2.Response {
	ctx := r.Context()
	members, err := discord.ListGuildMembers(ctx, 1000, 0)
	if err != nil {
		if e, ok := err.(*discord.Error); ok {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(e); err != nil {
				logger.JsonEncodeFailure(ctx, err)
			}
			return errors.NewErrorResponse(errors.Error{
				Code:    e.Code,
				Message: e.Error(),
				Data:    buf.Bytes(),
			})
		}
		return errors.NewInternalServerError()
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

	return http2.Response{Status: http2.StatusOk, GuildMembers: members}
})
