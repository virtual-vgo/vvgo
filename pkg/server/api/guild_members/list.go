package guild_members

import (
	"bytes"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/api/cache"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"time"
)

var HandleList = cache.Handle(4*3600*time.Second, func(r *http.Request) models.ApiResponse {
	ctx := r.Context()

	members, err := discord.ListGuildMembers(ctx, 1000, 0)
	if err != nil {
		if e, ok := err.(*discord.Error); ok {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(e); err != nil {
				logger.JsonEncodeFailure(ctx, err)
			}
			return http_helpers.NewErrorResponse(models.ApiError{
				Code:  e.Code,
				Error: e.Error(),
				Data:  buf.Bytes(),
			})
		}
		return http_helpers.NewInternalServerError()
	}
	return models.ApiResponse{Status: models.StatusOk, GuildMembers: members}
})
