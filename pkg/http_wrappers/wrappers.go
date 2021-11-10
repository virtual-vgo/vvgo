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
		}
		handler.ServeHTTP(writer, r.Clone(ctx))

		requestMetrics := traces.NewHttpRequestMetrics(r)
		responseMetrics := traces.NewHttpResponseMetrics(writer.code, writer.size)
		if trace != nil {
			redis.WriteSpan(
				trace.Finish().
					WithHttpRequestMetrics(requestMetrics).
					WithHttpResponseMetrics(responseMetrics).
					WithError(err),
			)
		}

		switch {
		case writer.code >= 200 && writer.code < 400:
			logger.
				WithFields(requestMetrics.Fields()).
				WithFields(responseMetrics.Fields()).
				Info("http server: request completed")
		default:
			logger.
				WithFields(requestMetrics.Fields()).
				WithFields(responseMetrics.Fields()).
				Warn("http server: request completed with error status")
		}
	}
}

func DoRequest(r *http.Request) (*http.Response, error) {
	var resp *http.Response
	var respErr error

	defer debugRequestOut(r)
	defer debugResponse(resp)

	span, spanOk := traces.NewSpanFromContext(r.Context(), "outgoing http request")
	if !spanOk {
		logger.Warn("http client: invalid trace context")
	}

	resp, respErr = http.DefaultClient.Do(r)
	requestMetrics := traces.NewHttpRequestMetrics(r)
	responseMetrics := traces.NewHttpResponseMetrics(resp.StatusCode, resp.ContentLength)
	if spanOk {
		redis.WriteSpan(
			span.Finish().
				WithHttpRequestMetrics(requestMetrics).
				WithHttpResponseMetrics(responseMetrics).
				WithError(respErr),
		)
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 400:
		logger.
			WithFields(requestMetrics.Fields()).
			WithFields(responseMetrics.Fields()).
			Info("http client: request completed")
	default:
		logger.
			WithFields(requestMetrics.Fields()).
			WithFields(responseMetrics.Fields()).
			Warn("http client: request completed with error status")
	}
	return resp, respErr
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
