package cache

import (
	"bytes"
	"context"
	"encoding/json"
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"strconv"
	"time"
)

func Handle(expires time.Duration, handler func(*http.Request) http2.Response) func(r *http.Request) http2.Response {
	return func(r *http.Request) http2.Response {
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

func readCache(ctx context.Context, key string) (http2.Response, bool) {
	var cacheRespJSON bytes.Buffer
	if err := redis.Do(ctx, redis.Cmd(&cacheRespJSON, "GET", key)); err != nil {
		logger.RedisFailure(ctx, err)
		return http2.Response{}, false
	}

	if cacheRespJSON.Len() != 0 {
		var cacheResp http2.Response
		if err := json.NewDecoder(&cacheRespJSON).Decode(&cacheResp); err != nil {
			logger.JsonDecodeFailure(ctx, err)
		} else {
			return cacheResp, true
		}
	}
	return http2.Response{}, false
}

func writeCache(ctx context.Context, key string, expires time.Duration, resp http2.Response) {
	if resp.Status != http2.StatusOk {
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
