package api

import (
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
	"net/http/pprof"
)

var logger = log.Logger()

var PublicFiles = "public"

type ServerConfig struct {
	ListenAddress    string
	MaxContentLength int64
	BasicAuthUser    string
	BasicAuthPass    string
}

func NewServer(config ServerConfig, sheets sheets.Sheets) *http.Server {
	mux := funcName(config, sheets)

	return &http.Server{
		Addr:     config.ListenAddress,
		Handler:  mux,
		ErrorLog: log.StdLogger(),
	}
}

func funcName(config ServerConfig, sheets sheets.Sheets) *http.ServeMux {
	auth := make(basicAuth)
	if config.BasicAuthUser != "" {
		auth[config.BasicAuthUser] = config.BasicAuthPass
	}

	mux := http.NewServeMux()

	// debug endpoints from net/http/pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("/sheets", auth.Authenticate(SheetsHandler{sheets}))
	mux.Handle("/sheets/", http.RedirectHandler("/sheets", http.StatusMovedPermanently))

	downloadHandler := DownloadHandler{
		SheetsBucketName: sheets.Bucket.DownloadURL,
	}
	mux.Handle("/download", auth.Authenticate(downloadHandler))
	mux.Handle("/upload", auth.Authenticate(UploadHandler{sheets}))
	mux.Handle("/version", http.HandlerFunc(Version))
	mux.Handle("/", http.FileServer(http.Dir("public")))
	return mux
}
