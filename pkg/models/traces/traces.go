package traces

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"math/rand"
	"net/http"
	"time"
)

const SpansRedisKey = "traces:spans"
const NextTraceIdRedisKey = "traces:next_trace_id"
const SpanContextKey = "trace_id"

type Waterfall struct {
	Span
	Children []Waterfall `json:"children,omitempty"`
}

func NewWaterfall(traceId uint64, spans []Span) (Waterfall, error) {
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
				return Waterfall{}, errors.New("multiple root spans detected")
			}
			root = &waterfalls[i]
		}
	}
	if root == nil {
		return Waterfall{}, errors.New("root span is missing")
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
			log.
				WithField("span_id", waterfalls[i].Id).
				WithField("trace_id", waterfalls[i].TraceId).
				Warn("span is orphaned")
			parent = root
		}
		parent.Children = append(parent.Children, waterfalls[i])
	}

	return *root, nil
}

type Span struct {
	Id           uint64           `json:"id"`
	Name         string           `json:"name"`
	TraceId      uint64           `json:"trace_id"`
	ParentId     uint64           `json:"parent_id"`
	StartTime    time.Time        `json:"start_time"`
	Duration     float64          `json:"duration"`
	HttpRequest  *HttpRequest     `json:"http_request,omitempty"`
	HttpResponse *HttpResponse    `json:"http_response,omitempty"`
	RedisQuery   *RedisQuery      `json:"redis_query,omitempty"`
	Error        string           `json:"error,omitempty"`
	ApiVersion   *version.Version `json:"api_version,omitempty"`
	ctx          context.Context
}

func NewTrace(ctx context.Context, id uint64, name string) Span {
	span := newSpan(nil, id, 0, name)
	span.ctx = context.WithValue(ctx, SpanContextKey, span)
	return span
}

func NewSpanFromContext(ctx context.Context, name string) (Span, bool) {
	parent, ok := ctx.Value(SpanContextKey).(Span)
	if !ok {
		return Span{}, false
	}
	return parent.NewSpan(name), true
}

func (x *Span) NewSpan(name string) Span {
	return newSpan(x.Context(), x.TraceId, x.Id, name)
}

func newSpan(ctx context.Context, traceId uint64, parentId uint64, name string) Span {
	return Span{
		Name:      name,
		TraceId:   traceId,
		ParentId:  parentId,
		Id:        rand.Uint64(),
		StartTime: time.Now(),
		ctx:       ctx,
	}
}

func (x Span) WithApiVersion() Span {
	v := version.Get()
	x.ApiVersion = &v
	return x
}

func (x Span) WithHttpRequest(r *http.Request) Span {
	x.HttpRequest = &HttpRequest{
		Method:    r.Method,
		Host:      r.URL.Host,
		Bytes:     r.ContentLength,
		Url:       r.URL.Path,
		UserAgent: r.UserAgent(),
	}
	return x
}

func (x Span) WithHttpResponse(code int, size int64) Span {
	x.HttpResponse = &HttpResponse{
		Code:  code,
		Bytes: size,
	}
	return x
}

func (x Span) WithRedisQuery(cmd string, args []string) Span {
	var argBytes int
	for _, arg := range args {
		argBytes += len(arg)
	}
	x.RedisQuery = &RedisQuery{
		Cmd:      cmd,
		ArgCount: len(args),
		ArgBytes: argBytes,
	}
	return x
}

func (x Span) WithError(err error) Span {
	if err != nil {
		x.Error = err.Error()
	}
	return x
}

func (x Span) IsHeadless() bool              { return x.TraceId == 0 }
func (x Span) Context() context.Context      { return x.ctx }
func (x Span) Start() Span                   { x.StartTime = time.Now(); x.Duration = 0; return x }
func (x Span) Finish() Span                  { x.Duration = time.Since(x.StartTime).Seconds(); return x }
func (x Span) FinishedAt(end time.Time) Span { x.Duration = end.Sub(x.StartTime).Seconds(); return x }

type HttpRequest struct {
	Host      string `json:"host"`
	Method    string `json:"method"`
	Bytes     int64  `json:"bytes"`
	Url       string `json:"url"`
	UserAgent string `json:"user_agent"`
}

type HttpResponse struct {
	Code  int   `json:"code"`
	Bytes int64 `json:"size"`
}

type RedisQuery struct {
	Cmd      string `json:"cmd"`
	ArgCount int    `json:"arg_count"`
	ArgBytes int    `json:"arg_bytes"`
}
