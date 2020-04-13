package api

import (
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net/http"
	"net/http/pprof"
)

const (
	SheetsBucketName = "sheets"
	ClixBucketName   = "clix"
	PartsBucketName  = "parts"
	PartsLockerName  = "parts.lock"
)

var logger = log.Logger()

var PublicFiles = "public"

type ServerConfig struct {
	ListenAddress    string
	MaxContentLength int64
	BasicAuthUser    string
	BasicAuthPass    string
	SheetsBucketName string
	ClixBucketName   string
	PartsBucketName  string
	PartsLockerKey   string
}

type FileBucket interface {
	PutFile(file *storage.File) bool
	DownloadURL(name string) (string, error)
}

type Storage struct {
	parts.Parts
	Sheets FileBucket
	Clix   FileBucket
	ServerConfig
}

func NewStorage(client *storage.Client, config ServerConfig) *Storage {
	sheetsBucket := client.NewBucket(config.SheetsBucketName)
	clixBucket := client.NewBucket(config.ClixBucketName)
	partsBucket := client.NewBucket(config.PartsBucketName)
	partsLocker := client.NewLocker(config.PartsLockerKey)
	if sheetsBucket == nil || clixBucket == nil || partsBucket == nil || partsLocker == nil {
		return nil
	}

	return &Storage{
		Parts: parts.Parts{
			Bucket: partsBucket,
			Locker: partsLocker,
		},
		Sheets:       sheetsBucket,
		Clix:         clixBucket,
		ServerConfig: config,
	}
}

func NewServer(config ServerConfig, database *Storage) *http.Server {
	auth := make(basicAuth)
	if config.BasicAuthUser != "" {
		auth[config.BasicAuthUser] = config.BasicAuthPass
	}

	mux := http.NewServeMux()

	mux.Handle("/auth",
		auth.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("authenticated"))
		})),
	)

	// debug endpoints from net/http/pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("/parts", auth.Authenticate(PartsHandler{database}))
	mux.Handle("/parts/", http.RedirectHandler("/parts", http.StatusMovedPermanently))

	downloadHandler := DownloadHandler{
		SheetsBucketName: database.Sheets.DownloadURL,
		ClixBucketName:   database.Clix.DownloadURL,
	}
	mux.Handle("/download", auth.Authenticate(downloadHandler))

	uploadHandler := UploadHandler{database}
	mux.Handle("/upload", auth.Authenticate(uploadHandler))

	mux.Handle("/version", http.HandlerFunc(Version))
	mux.Handle("/", http.FileServer(http.Dir("public")))

	return &http.Server{
		Addr:     config.ListenAddress,
		Handler:  mux,
		ErrorLog: log.StdLogger(),
	}
}
