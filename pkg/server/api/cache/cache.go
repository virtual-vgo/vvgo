package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"net/http"
	"strconv"
	"time"
)

func Handle(expires time.Duration, handler func(*http.Request) models.ApiResponse) func(r *http.Request) models.ApiResponse {
	return func(r *http.Request) models.ApiResponse {
		ctx := r.Context()
		key := "response_cache:" + r.URL.String()
		if response, ok := readCache(ctx, key); ok {
			return response
		}
		logger.Info("cache miss", key)

		resp := handler(r)
		writeCache(ctx, key, expires, resp)
		return resp
	}
}

func readCache(ctx context.Context, key string) (models.ApiResponse, bool) {
	var cacheRespJSON bytes.Buffer
	if err := redis.Do(ctx, redis.Cmd(&cacheRespJSON, "GET", key)); err != nil {
		logger.RedisFailure(ctx, err)
		return models.ApiResponse{}, false
	}

	if cacheRespJSON.Len() != 0 {
		var cacheResp models.ApiResponse
		if err := json.NewDecoder(&cacheRespJSON).Decode(&cacheResp); err != nil {
			logger.JsonDecodeFailure(ctx, err)
		} else {
			return cacheResp, true
		}
	}
	return models.ApiResponse{}, false
}

func writeCache(ctx context.Context, key string, expires time.Duration, resp models.ApiResponse) {
	var cacheRespJSON bytes.Buffer
	if err := json.NewEncoder(&cacheRespJSON).Encode(resp); err != nil {
		logger.JsonEncodeFailure(ctx, err)
		return
	}

	ttl := int(expires.Seconds())
	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", key, strconv.Itoa(ttl), cacheRespJSON.String())); err != nil {
		logger.RedisFailure(ctx, err)
		return
	}
}
