package guild_members

import (
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/cache"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"strconv"
	"time"
)

type SearchRequest struct {
	Limit int
	Query string
}

var HandleSearch = cache.Handle(60*time.Second, func(r *http.Request) api.Response {
	ctx := r.Context()

	queryParams := r.URL.Query()
	limit, _ := strconv.Atoi(queryParams.Get("limit"))
	params := SearchRequest{
		Limit: limit,
		Query: queryParams.Get("query"),
	}

	if params.Query == "" {
		return response.NewBadRequestError("query is required")
	}

	members, err := discord.SearchGuildMembers(ctx, params.Query, params.Limit)
	if err != nil {
		logger.MethodFailure(ctx, "discord.SearchGuildMembers", err)
		return response.NewInternalServerError()
	}

	saveGuildMembers(ctx, members)
	return api.Response{Status: api.StatusOk, GuildMembers: members}
})
