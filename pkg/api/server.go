package api

import (
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/storage"
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
	SheetsBucketName string
	ClixBucketName   string
	PartsBucketName  string
	PartsLockerKey   string
	PrepRepToken     string
	AdminToken       string
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
	members := BasicAuth{config.BasicAuthUser: config.BasicAuthPass}
	prepRep := TokenAuth{config.PrepRepToken, config.AdminToken}
	admin := TokenAuth{config.AdminToken}

	mux := http.NewServeMux()

	mux.Handle("/auth",
		members.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("authenticated"))
		})),
	)

	// debug endpoints from net/http/pprof
	pprofMux := http.NewServeMux()
	pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/", admin.Authenticate(pprofMux))

	mux.Handle("/parts", members.Authenticate(PartsHandler{database}))
	mux.Handle("/parts/", http.RedirectHandler("/parts", http.StatusMovedPermanently))

	downloadHandler := DownloadHandler{
		config.SheetsBucketName: database.Sheets.DownloadURL,
		config.ClixBucketName:   database.Clix.DownloadURL,
	}
	mux.Handle("/download", prepRep.Authenticate(downloadHandler))

	// Uploads
	uploadHandler := UploadHandler{database}
	mux.Handle("/upload", prepRep.Authenticate(uploadHandler))

	mux.Handle("/version", http.HandlerFunc(Version))
	mux.Handle("/", http.FileServer(http.Dir("public")))

	return &http.Server{
		Addr:     config.ListenAddress,
		Handler:  mux,
		ErrorLog: log.StdLogger(),
	}
}
