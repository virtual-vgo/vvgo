package api

import (
	"bytes"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"strconv"
	"time"
)

type GuildMembersRequest struct {
	Limit string
	Query string
}

var GuildMembers = CacheApiResponse(60*time.Second, func(r *http.Request) models.ApiResponse {
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

func CacheApiResponse(expires time.Duration, handler func(*http.Request) models.ApiResponse) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var respJSON bytes.Buffer
		key := "api_response_cache:" + r.URL.String()
		if err := redis.Do(ctx, redis.Cmd(&respJSON, "GET", key)); err != nil {
			logger.RedisFailure(ctx, err)
		}

		var resp models.ApiResponse
		if respJSON.Len() == 0 {
			logger.Info("cache miss", key)
			resp = handler(r)
		} else if err := json.NewDecoder(&respJSON).Decode(&resp); err != nil {
			logger.JsonDecodeFailure(ctx, err)
			resp = handler(r)
		} else {
			http_helpers.WriteAPIResponse(ctx, w, resp)
			return
		}

		respJSON.Reset()
		if err := json.NewEncoder(&respJSON).Encode(resp); err != nil {
			logger.JsonEncodeFailure(ctx, err)
			http_helpers.WriteInternalServerError(ctx, w)
			return
		}

		ttl := int(expires.Seconds())
		if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", key, strconv.Itoa(ttl), respJSON.String())); err != nil {
			logger.RedisFailure(ctx, err)
		}
		http_helpers.WriteAPIResponse(ctx, w, resp)
	}
}

func guildMembersHandler(r *http.Request) models.ApiResponse {
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
}
