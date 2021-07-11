package http_wrappers

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

var logger = log.Logger()

// middleware response writer that captures the http response code and other metrics
type responseWriter struct {
	code int
	size int
	http.ResponseWriter
}

func (x responseWriter) WriteHeader(code int) {
	x.code = code
	x.ResponseWriter.WriteHeader(code)
}

func (x responseWriter) Write(b []byte) (int, error) {
	version.SetVersionHeaders(x)
	n, err := x.ResponseWriter.Write(b)
	x.size += n
	return n, err
}

func Handler(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		writer := responseWriter{ResponseWriter: w, code: http.StatusOK} // this is the default status code
		debugRequestIn(r)
		handler.ServeHTTP(writer, r)

		// submit results
		logger.WithFields(logrus.Fields{
			"request_method":     r.Method,
			"request_size":       r.ContentLength,
			"request_path":       r.URL.Path,
			"request_host":       r.URL.Host,
			"request_user_agent": r.UserAgent(),
			"response_status":    writer.code,
			"response_size":      writer.size,
			"start_time":         start,
			"duration":           time.Since(start).Seconds(),
		}).Info("incoming http server request completed")
	}
}

func DoRequest(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// do the request
	debugRequestOut(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	debugResponse(resp)

	// submit results
	logger.WithFields(logrus.Fields{
		"request_method":  req.Method,
		"request_size":    req.ContentLength,
		"request_path":    req.URL.Path,
		"request_host":    req.URL.Host,
		"response_status": resp.StatusCode,
		"response_size":   resp.ContentLength,
		"start_time":      start,
		"duration":        time.Since(start).Seconds(),
	}).Info("outgoing http client request completed")
	return resp, err
}

func Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return DoRequest(req)
}

func PostForm(url string, data url.Values) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return DoRequest(req)
}

var DebugHTTP bool

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
