package api

import (
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"time"
)

type GuildMembersRequest struct {
	Limit string
	Query string
}

var GuildMembers = CacheResponse(60*time.Second, func(r *http.Request) models.ApiResponse {
	ctx := r.Context()

	queryParams := r.URL.Query()
	params := GuildMembersRequest{
		Limit: queryParams.Get("limit"),
		Query: queryParams.Get("query"),
	}

	if params.Query == "" {
		return http_helpers.NewBadRequestError("query is required")
	}

	guildMembers, err := discord.SearchGuildMembers(ctx, params.Query, params.Limit)
	if err != nil {
		logger.MethodFailure(ctx, "discord.SearchGuildMembers", err)
		return http_helpers.NewInternalServerError()
	}

	return models.ApiResponse{
		Status:       models.StatusOk,
		GuildMembers: guildMembers,
	}
})
