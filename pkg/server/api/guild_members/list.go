package guild_members

import (
	"bytes"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/api/cache"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"strconv"
	"time"
)

type ListRequest struct {
	Limit int
	After int
}

var HandleList = cache.Handle(60*time.Second, func(r *http.Request) models.ApiResponse {
	ctx := r.Context()

	queryParams := r.URL.Query()
	limit, _ := strconv.Atoi(queryParams.Get("limit"))
	after, _ := strconv.Atoi(queryParams.Get("after"))
	params := ListRequest{
		Limit: limit,
		After: after,
	}

	members, err := discord.ListGuildMembers(ctx, params.Limit, params.After)
	if err != nil {
		if e, ok := err.(*discord.Error); ok {
			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(e)
			return http_helpers.NewErrorResponse(models.ApiError{
				Code:  e.Code,
				Error: e.Error(),
				Data:  buf.Bytes(),
			})
		}
		return http_helpers.NewInternalServerError()
	}

	saveGuildMembers(ctx, members)
	return models.ApiResponse{Status: models.StatusOk, GuildMembers: members}
})
