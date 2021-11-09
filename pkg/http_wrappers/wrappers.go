package http_wrappers

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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
		start := time.Now()

		writer := ResponseWriter{ResponseWriter: w, code: http.StatusOK} // this is the default status code
		debugRequestIn(r)
		handler.ServeHTTP(&writer, r)
		entry := apilog.Entry{
			Request: apilog.Request{
				Method: r.Method,
				Size:   r.ContentLength,
				Url:    apilog.Url{Path: r.URL.Path, Host: r.URL.Host},
			},
			Response: apilog.Response{
				Code: writer.code,
				Size: int64(writer.size),
			},
			Duration: time.Since(start),
		}

		fields := log.Fields{
			"request":  entry.Request,
			"response": entry.Response,
			"duration": entry.Duration,
		}

		// submit results
		switch {
		case writer.code >= 200 && writer.code < 400:
			logger.WithFields(fields).Info("http server: request completed")
		default:
			logger.WithFields(fields).Error("http server: request completed with error")
		}
	}
}

func DoRequest(req *http.Request) (*http.Response, error) {
	start := time.Now()

	debugRequestOut(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	debugResponse(resp)

	logger.WithFields(log.Fields{
		"request": apilog.Request{
			Method: req.Method,
			Size:   req.ContentLength,
			Url:    apilog.Url{Path: req.URL.Path, Host: req.URL.Host},
		},
		"response": apilog.Response{
			Code: resp.StatusCode,
			Size: resp.ContentLength,
		},
		"duration": time.Since(start).Seconds(),
	}).Info("http client request completed")
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
