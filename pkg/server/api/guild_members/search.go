package guild_members

import (
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/api/cache"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"time"
)

type SearchRequest struct {
	Limit string
	Query string
}

var HandleSearch = cache.Handle(60*time.Second, func(r *http.Request) models.ApiResponse {
	ctx := r.Context()

	queryParams := r.URL.Query()
	params := SearchRequest{
		Limit: queryParams.Get("limit"),
		Query: queryParams.Get("query"),
	}

	if params.Query == "" {
		return http_helpers.NewBadRequestError("query is required")
	}

	members, err := discord.SearchGuildMembers(ctx, params.Query, params.Limit)
	if err != nil {
		logger.MethodFailure(ctx, "discord.SearchGuildMembers", err)
		return http_helpers.NewInternalServerError()
	}

	saveGuildMembers(ctx, members)
	return models.ApiResponse{Status: models.StatusOk, GuildMembers: members}
})
