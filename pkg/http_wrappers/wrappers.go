package http_wrappers

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models/apilog"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
	"net/http/httputil"
	"time"
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
	size int
	http.ResponseWriter
}

func (x *ResponseWriter) WriteHeader(code int) {
	x.code = code
	x.ResponseWriter.WriteHeader(code)
}

func (x *ResponseWriter) Write(b []byte) (int, error) {
	version.SetVersionHeaders(x)
	n, err := x.ResponseWriter.Write(b)
	x.size += n
	return n, err
}

func Handler(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writer := ResponseWriter{ResponseWriter: w, code: http.StatusOK}
		start := time.Now()
		handler.ServeHTTP(&writer, r)
		entry := apilog.Entry{
			StartTime:        start,
			RequestMethod:    r.Method,
			RequestHost:      r.URL.Host,
			RequestBytes:     r.ContentLength,
			RequestUrl:       r.URL.Path,
			RequestUserAgent: r.UserAgent(),
			ResponseCode:     writer.code,
			ResponseBytes:    int64(writer.size),
			DurationSeconds:  time.Since(start).Seconds(),
		}
		if err := redis.WriteLog(r.Context(), entry); err != nil {
			logger.RedisFailure(r.Context(), err)
		}

		switch {
		case writer.code >= 200 && writer.code < 400:
			logger.WithFields(entry.Fields()).Info("http server: request completed")
		default:
			logger.WithFields(entry.Fields()).Error("http server: request completed with non-200 status")
		}
		debugRequestIn(r)

	}
}

func DoRequest(r *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	entry := apilog.Entry{
		StartTime:       start,
		RequestMethod:   r.Method,
		RequestHost:     r.URL.Host,
		RequestBytes:    r.ContentLength,
		RequestUrl:      r.URL.Path,
		ResponseCode:    resp.StatusCode,
		ResponseBytes:   resp.ContentLength,
		DurationSeconds: time.Since(start).Seconds(),
	}
	if err := redis.WriteLog(r.Context(), entry); err != nil {
		logger.RedisFailure(r.Context(), err)
	}
	logger.WithFields(entry.Fields()).Info("http client: request completed")
	debugRequestOut(r)
	debugResponse(resp)

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
