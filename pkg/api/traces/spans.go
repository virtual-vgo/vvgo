package traces

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"strings"
	"time"
)

func ServeSpans(r *http.Request) api.Response {
	ctx := r.Context()

	var data UrlQueryParams
	data.ReadParams(r.URL.Query())

	spans, err := ListSpans(ctx, data.End, data.Start)
	if err != nil {
		logger.RedisFailure(ctx, err)
		return errors.NewRedisError(err)
	}
	if len(spans) > data.Limit {
		spans = spans[:data.Limit]
	}

	return api.Response{Status: api.StatusOk, Spans: spans}
}

func ListSpans(ctx context.Context, start, end time.Time) ([]tracing.Span, error) {
	startString := fmt.Sprintf("%f", time.Duration(start.UnixNano()).Seconds())
	endString := fmt.Sprintf("%f", time.Duration(end.UnixNano()).Seconds())

	cmd := redis.ZRANGEBYSCORE
	if end.Before(start) {
		cmd = redis.ZREVRANGEBYSCORE
	}

	var entriesJSON []string
	if err := redis.Do(ctx, redis.Cmd(&entriesJSON, cmd, tracing.SpansRedisKey, startString, endString)); err != nil {
		return nil, err
	}
	spans := make([]tracing.Span, 0, len(entriesJSON))
	for _, logJSON := range entriesJSON {
		var entry tracing.Span
		if err := json.NewDecoder(strings.NewReader(logJSON)).Decode(&entry); err != nil {
			logger.WithError(err).Error("json.Decode() failed")
		}
		spans = append(spans, entry)
	}
	return spans, nil
}
