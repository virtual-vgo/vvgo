package tracing

import (
	"context"
	"fmt"
	"github.com/honeycombio/beeline-go"
	"github.com/honeycombio/beeline-go/trace"
	"github.com/honeycombio/beeline-go/wrappers/hnynethttp"
	"net/http"
)

type Config struct {
	ServiceName       string `split_words:"true" default:"vvgo"`
	HoneycombWriteKey string `split_words:"true" default:""`
	HoneycombDataset  string `split_words:"true" default:"development"`
}

func Initialize(config Config) {
	beeline.Init(beeline.Config{
		ServiceName: config.ServiceName,
		WriteKey:    config.HoneycombWriteKey,
		Dataset:     config.HoneycombDataset,
	})
}

func Close() {
	beeline.Close()
}

type Span struct {
	*trace.Span
}

func StartSpan(ctx context.Context, name string) (context.Context, Span) {
	ctx, honeycombSpan := beeline.StartSpan(ctx, name)
	return ctx, Span{honeycombSpan}
}

func WrapHandler(handler http.Handler) http.Handler {
	return hnynethttp.WrapHandler(handler)
}
func AddError(ctx context.Context, err error) { AddField(ctx, "error", err) }
func AddField(ctx context.Context, key string, val interface{}) {
	beeline.AddField(ctx, key, val)
}

func DoHttpRequest(r *http.Request) (*http.Response, error) {
	_, span := StartSpan(r.Context(), fmt.Sprintf("%s: %s", r.Method, r.URL.String()))
	defer span.Send()
	span.AddField("method", r.Method)
	span.AddField("uri", r.URL.String())
	span.AddField("content_length", r.ContentLength)
	return http.DefaultClient.Do(r)
}
