package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"strconv"
	"time"
)

func Handle(expires time.Duration, handler func(*http.Request) api.Response) func(r *http.Request) api.Response {
	return func(r *http.Request) api.Response {
		ctx := r.Context()
		key := "response_cache:" + r.URL.String()
		if cachedResp, ok := readCache(ctx, key); ok {
			return cachedResp
		}
		logger.WithField("cache_key", key).Info("cache miss")
		resp := handler(r)
		writeCache(ctx, key, expires, resp)
		return resp
	}
}

func readCache(ctx context.Context, key string) (api.Response, bool) {
	var cacheRespJSON bytes.Buffer
	if err := redis.Do(ctx, redis.Cmd(&cacheRespJSON, "GET", key)); err != nil {
		logger.RedisFailure(ctx, err)
		return api.Response{}, false
	}

	if cacheRespJSON.Len() != 0 {
		var cacheResp api.Response
		if err := json.NewDecoder(&cacheRespJSON).Decode(&cacheResp); err != nil {
			logger.JsonDecodeFailure(ctx, err)
		} else {
			return cacheResp, true
		}
	}
	return api.Response{}, false
}

func writeCache(ctx context.Context, key string, expires time.Duration, resp api.Response) {
	if resp.Status != api.StatusOk {
		return
	}

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
