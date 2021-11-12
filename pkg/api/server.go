package api

import (
	"fmt"
	logurs "github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"log"
	"net/http"
	"net/http/httputil"
)

type Server struct {
	*http.Server
}

func NewServer(listenAddress string) *Server {
	return &Server{
		Server: &http.Server{
			Addr:     listenAddress,
			Handler:  ServeHTTP(Routes()),
			ErrorLog: log.New(logurs.New().Writer(), "", 0),
		},
	}
}

type ResponseWriter struct {
	Code       int
	CountBytes int64
	http.ResponseWriter
}

func (x *ResponseWriter) WriteHeader(code int) {
	x.Code = code
	x.ResponseWriter.WriteHeader(code)
}

func (x *ResponseWriter) Write(b []byte) (int, error) {
	version.SetVersionHeaders(x)
	n, err := x.ResponseWriter.Write(b)
	x.CountBytes += int64(n)
	return n, err
}

func ServeHTTP(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if config.Env.DebugHTTP {
				recvBytes, _ := httputil.DumpRequest(r, true)
				fmt.Println("--- BEGIN DEBUG HTTP SERVER REQUEST ---")
				fmt.Println("Received request:")
				fmt.Println(string(recvBytes))
				fmt.Println("--- END DEBUG HTTP SERVER REQUEST ---")
			}
		}()

		ctx := r.Context()
		writer := &ResponseWriter{ResponseWriter: w, Code: http.StatusOK}
		trace, err := tracing.NewTrace(ctx, "incoming http request")
		if err != nil {
			logger.RedisFailure(ctx, err)
		} else {
			ctx = trace.Context()
		}
		handler.ServeHTTP(writer, r.Clone(ctx))

		requestMetrics := tracing.NewHttpRequestMetrics(r)
		responseMetrics := tracing.NewHttpResponseMetrics(writer.Code, writer.CountBytes)
		if trace != nil {
			tracing.WriteSpan(
				trace.Finish().
					WithHttpRequestMetrics(requestMetrics).
					WithHttpResponseMetrics(responseMetrics).
					WithError(err),
			)
		}

		switch {
		case writer.Code >= 200 && writer.Code < 400:
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
