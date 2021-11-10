package traces

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/logger"
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
	Id           uint64               `json:"id"`
	Name         string               `json:"name"`
	TraceId      uint64               `json:"trace_id"`
	ParentId     uint64               `json:"parent_id"`
	StartTime    time.Time            `json:"start_time"`
	Duration     float64              `json:"duration"`
	HttpRequest  *HttpRequestMetrics  `json:"http_request,omitempty"`
	HttpResponse *HttpResponseMetrics `json:"http_response,omitempty"`
	RedisQuery   *RedisQueryMetrics   `json:"redis_query,omitempty"`
	Error        string               `json:"error,omitempty"`
	ApiVersion   *version.Version     `json:"api_version,omitempty"`
	ctx          context.Context
}

func NewTrace(ctx context.Context, id uint64, name string) Span {
	span := newSpan(nil, id, 0, name)
	span.ctx = context.WithValue(ctx, SpanContextKey, &span)
	return span
}

func NewSpanFromContext(ctx context.Context, name string) (Span, bool) {
	parent, ok := ctx.Value(SpanContextKey).(*Span)
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

func (x Span) WithHttpRequestMetrics(metrics HttpRequestMetrics) Span {
	x.HttpRequest = &metrics
	return x
}

func (x Span) WithHttpResponseMetrics(metrics HttpResponseMetrics) Span {
	x.HttpResponse = &metrics
	return x
}

func (x Span) WithRedisQuery(metrics RedisQueryMetrics) Span {
	x.RedisQuery = &metrics
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

type HttpRequestMetrics struct {
	Host      string `json:"host"`
	Method    string `json:"method"`
	Bytes     int64  `json:"bytes"`
	Url       string `json:"url"`
	UserAgent string `json:"user_agent"`
}

func NewHttpRequestMetrics(r *http.Request) HttpRequestMetrics {
	return HttpRequestMetrics{
		Method:    r.Method,
		Host:      r.URL.Host,
		Bytes:     r.ContentLength,
		Url:       r.URL.Path,
		UserAgent: r.UserAgent(),
	}
}

func (x HttpRequestMetrics) Fields() map[string]interface{} { return fieldsFromStruct(x) }

type HttpResponseMetrics struct {
	Code  int   `json:"code"`
	Bytes int64 `json:"size"`
}

func NewHttpResponseMetrics(code int, size int64) HttpResponseMetrics {
	return HttpResponseMetrics{
		Code:  code,
		Bytes: size,
	}
}

func (x HttpResponseMetrics) Fields() map[string]interface{} { return fieldsFromStruct(x) }

type RedisQueryMetrics struct {
	Cmd      string `json:"cmd"`
	ArgCount int    `json:"arg_count"`
	ArgBytes int    `json:"arg_bytes"`
}

func NewRedisQueryMetrics(cmd string, args []string) RedisQueryMetrics {
	var argBytes int
	for _, arg := range args {
		argBytes += len(arg)
	}
	return RedisQueryMetrics{
		Cmd:      cmd,
		ArgCount: len(args),
		ArgBytes: argBytes,
	}
}

func (x RedisQueryMetrics) Fields() map[string]interface{} { return fieldsFromStruct(x) }

func fieldsFromStruct(str interface{}) map[string]interface{} {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(str); err != nil {
		logger.JsonEncodeFailure(context.Background(), err)
		return nil
	}

	var fields map[string]interface{}
	if err := json.NewDecoder(&buf).Decode(&fields); err != nil {
		logger.JsonDecodeFailure(context.Background(), err)
		return nil
	}
	return fields
}
