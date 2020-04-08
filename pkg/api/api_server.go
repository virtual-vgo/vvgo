package api

import (
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"
)

var logger = log.Logger()

var Public = "public"

type Config struct {
	MaxContentLength int64
	BasicAuthUser    string
	BasicAuthPass    string
}

type Server struct {
	Config
	*http.ServeMux
	storage.ObjectStorage
	basicAuth
}

func NewServer(store storage.ObjectStorage, config Config) *Server {
	auth := make(basicAuth)
	if config.BasicAuthUser != "" {
		auth[config.BasicAuthUser] = config.BasicAuthPass
	}
	server := Server{
		ObjectStorage: store,
		Config:        config,
		ServeMux:      http.NewServeMux(),
		basicAuth:     auth,
	}

	// debug endpoints from net/http/pprof
	server.HandleFunc("/debug/pprof/", pprof.Index)
	server.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	server.HandleFunc("/debug/pprof/profile", pprof.Profile)
	server.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	server.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server.Handle("/sheets", auth.Authenticate(server.SheetsIndex))
	server.Handle("/sheets/", http.RedirectHandler("/sheets", http.StatusMovedPermanently))
	server.Handle("/sheets/upload", auth.Authenticate(server.SheetsUpload))
	server.Handle("/download", auth.Authenticate(server.Download))
	server.Handle("/version", HandlerFunc(server.Version))
	server.Handle("/", http.FileServer(http.Dir("public")))
	return &server
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// This is http.ResponseWriter middleware that captures the response code
// and other info that might useful in logs
type responseWriter struct {
	code int
	http.ResponseWriter
}

func (x *responseWriter) WriteHeader(code int) {
	x.code = code
	x.ResponseWriter.WriteHeader(code)
}

func (handlerFunc HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	results := responseWriter{ResponseWriter: w}
	handlerFunc(&results, r)

	clientIP := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
	if clientIP == "" {
		clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	fields := logrus.Fields{
		"client_ip":       clientIP,
		"request_path":    r.URL.EscapedPath(),
		"user_agent":      r.UserAgent(),
		"request_method":  r.Method,
		"request_size":    r.ContentLength,
		"request_seconds": time.Since(start).Seconds(),
		"status_code":     results.code,
	}
	switch true {
	case results.code >= 500:
		logger.WithFields(fields).Error("request failed")
	case results.code >= 400:
		logger.WithFields(fields).Error("invalid request")
	default:
		logger.WithFields(fields).Info("request completed")
	}
}
