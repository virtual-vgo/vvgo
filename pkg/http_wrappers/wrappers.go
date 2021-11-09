package http_wrappers

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models/traces"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
	"net/http/httputil"
)

var DebugHTTP = false

func NoFollow(client *http.Client) *http.Client {
	if client == nil {
		client = new(http.Client)
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return client
}

// ResponseWriter middleware response writer that captures the http response code and other metrics
type ResponseWriter struct {
	code int
	size int64
	http.ResponseWriter
}

func (x *ResponseWriter) WriteHeader(code int) {
	x.code = code
	x.ResponseWriter.WriteHeader(code)
}

func (x *ResponseWriter) Write(b []byte) (int, error) {
	version.SetVersionHeaders(x)
	n, err := x.ResponseWriter.Write(b)
	x.size += int64(n)
	return n, err
}

func Handler(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer debugRequestIn(r)

		ctx := r.Context()
		writer := &ResponseWriter{ResponseWriter: w, code: http.StatusOK}
		trace, err := redis.NewTrace(ctx, "incoming http request")
		if err != nil {
			logger.RedisFailure(ctx, err)
		} else {
			ctx = trace.Context()
			defer redis.WriteSpan(trace.
				WithHttpRequest(r).
				WithHttpResponse(writer.code, writer.size),
			)
		}
		handler.ServeHTTP(writer, r.Clone(ctx))

		switch {
		case writer.code >= 200 && writer.code < 400:
			logger.
				WithField("response_code", writer.code).
				WithField("request_url", r.URL.Path).
				Info("http server: request completed")
		default:
			logger.
				WithField("response_code", writer.code).
				WithField("request_url", r.URL.Path).
				Warn("http server: request completed with error status")
		}
	}
}

func DoRequest(r *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	defer debugRequestOut(r)
	defer debugResponse(resp)

	span, ok := traces.NewSpanFromContext(r.Context(), "outgoing http request")
	if !ok {
		logger.Warn("http client: invalid trace context")
	} else {
		defer redis.WriteSpan(
			span.WithHttpRequest(r).
				WithHttpResponse(resp.StatusCode, resp.ContentLength).
				WithError(err),
		)
	}

	resp, err = http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 400:
		logger.
			WithField("response_code", resp.StatusCode).
			WithField("request_url", r.URL.Path).
			Info("http client: request completed")
	default:
		logger.
			WithField("response_code", resp.StatusCode).
			WithField("request_url", r.URL.Path).
			Warn("http client: request completed with error status")
	}
	return resp, err
}

func Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return DoRequest(req)
}

func debugRequestOut(r *http.Request) {
	if DebugHTTP {
		fmt.Println("sending request:")
		buf, _ := httputil.DumpRequestOut(r, true)
		fmt.Println(string(buf))
	}
}

func debugRequestIn(r *http.Request) {
	if DebugHTTP {
		fmt.Println("received request:")
		buf, _ := httputil.DumpRequest(r, true)
		fmt.Println(string(buf))
	}
}

func debugResponse(r *http.Response) {
	if DebugHTTP {
		fmt.Println("received response:")
		buf, _ := httputil.DumpResponse(r, true)
		fmt.Println(string(buf))
	}
}
