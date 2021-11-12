package traces

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"sort"
)

type Waterfall struct {
	tracing.Span
	Children []Waterfall `json:"children,omitempty"`
}

var ErrMultipleRootSpans = errors.New("multiple root spans detected")
var ErrRootSpanIsMissing = errors.New("root span is missing")

func NewWaterfall(traceId uint64, spans []tracing.Span) (Waterfall, error) {
	var waterfalls []Waterfall
	for i := range spans {
		if spans[i].TraceId == traceId {
			waterfalls = append(waterfalls, Waterfall{Span: spans[i]})
		}
	}

	var root *Waterfall
	for i := range waterfalls {
		if waterfalls[i].ParentId == 0 {
			if root != nil {
				return Waterfall{}, ErrMultipleRootSpans
			}
			root = &waterfalls[i]
		}
	}
	if root == nil {
		return Waterfall{}, ErrRootSpanIsMissing
	}

	spanIdToWaterfall := make(map[uint64]*Waterfall, len(waterfalls))
	for i := range waterfalls {
		spanIdToWaterfall[waterfalls[i].Id] = &waterfalls[i]
	}

	for i := range waterfalls {
		if waterfalls[i].ParentId == 0 {
			continue
		}

		parent, ok := spanIdToWaterfall[waterfalls[i].ParentId]
		if !ok {
			logrus.
				WithField("span_id", waterfalls[i].Id).
				WithField("trace_id", waterfalls[i].TraceId).
				Warn("span is orphaned")
			parent = root
		}
		parent.Children = append(parent.Children, waterfalls[i])
	}

	return *root, nil
}

func ServeWaterfall(r *http.Request) api.Response {
	ctx := r.Context()
	var data UrlQueryParams
	data.ReadParams(r.URL.Query())

	spans, err := ListSpans(ctx, data.Start, data.End)
	if err != nil {
		logger.RedisFailure(ctx, err)
		return response.NewRedisError(err)
	}

	traceIdsSet := make(map[uint64]struct{})
	for i := range spans {
		traceIdsSet[spans[i].TraceId] = struct{}{}
	}

	var thisTraceId uint64
	if span, ok := tracing.NewSpanFromContext(ctx, "throwaway"); ok {
		thisTraceId = span.TraceId
	}
	delete(traceIdsSet, thisTraceId)

	traceIds := make([]uint64, 0, len(traceIdsSet))
	for id := range traceIdsSet {
		traceIds = append(traceIds, id)
	}
	sort.Slice(traceIds, func(i, j int) bool { return traceIds[i] > traceIds[j] })

	waterfalls := make([]Waterfall, 0, data.Limit)
	for _, traceId := range traceIds {
		waterfall, err := NewWaterfall(traceId, spans)
		if err != nil {
			logger.WithField("trace_id", traceId).MethodFailure(ctx, "traces.NewWaterfall", err)
			continue
		}
		waterfalls = append(waterfalls, waterfall)

		if len(waterfalls) > data.Limit {
			break
		}
	}

	return api.Response{Status: api.StatusOk, Waterfalls: waterfalls[:]}
}
