package traces

import (
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/models/traces"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

type Request struct {
	Start time.Time
	End   time.Time
	Limit int
}

func (x *Request) ReadParams(params url.Values) {
	x.Start, _ = time.Parse(time.RFC3339, params.Get("start"))
	if x.Start.IsZero() {
		x.Start = time.Now().Add(-52 * 7 * 24 * 3600 * time.Second)
	}

	x.End, _ = time.Parse(time.RFC3339, params.Get("end"))
	if x.End.IsZero() {
		x.End = time.Now()
	}

	x.Limit, _ = strconv.Atoi(params.Get("limit"))
	if x.Limit == 0 {
		x.Limit = 1
	}
}

func HandleSpans(r *http.Request) models.ApiResponse {
	ctx := r.Context()

	var data Request
	data.ReadParams(r.URL.Query())

	spans, err := redis.ListSpans(ctx, data.End, data.Start)
	if err != nil {
		logger.RedisFailure(ctx, err)
		return http_helpers.NewRedisError(err)
	}
	if len(spans) > data.Limit {
		spans = spans[:data.Limit]
	}

	return models.ApiResponse{Status: models.StatusOk, Spans: spans}
}

func HandleWaterfall(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	var data Request
	data.ReadParams(r.URL.Query())

	spans, err := redis.ListSpans(ctx, data.Start, data.End)
	if err != nil {
		logger.RedisFailure(ctx, err)
		return http_helpers.NewRedisError(err)
	}

	traceIdsSet := make(map[uint64]struct{})
	for i := range spans {
		traceIdsSet[spans[i].TraceId] = struct{}{}
	}

	var thisTraceId uint64
	if span, ok := traces.NewSpanFromContext(ctx, "throwaway"); ok {
		thisTraceId = span.TraceId
	}
	delete(traceIdsSet, thisTraceId)

	traceIds := make([]uint64, 0, len(traceIdsSet))
	for id := range traceIdsSet {
		traceIds = append(traceIds, id)
	}
	sort.Slice(traceIds, func(i, j int) bool { return traceIds[i] > traceIds[j] })

	waterfalls := make([]traces.Waterfall, 0, data.Limit)
	for _, traceId := range traceIds {
		waterfall, err := traces.NewWaterfall(traceId, spans)
		if err != nil {
			logger.WithField("trace_id", traceId).MethodFailure(ctx, "traces.NewWaterfall", err)
			continue
		}
		waterfalls = append(waterfalls, waterfall)

		if len(waterfalls) > data.Limit {
			break
		}
	}

	return models.ApiResponse{Status: models.StatusOk, Waterfalls: waterfalls[:]}
}
