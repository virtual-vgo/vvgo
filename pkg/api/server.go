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
	ListenAddress    string `split_words:"true" default:"0.0.0.0:8080"`
	MaxContentLength int64  `split_words:"true" default:"10000000"`
	SheetsBucketName string `split_words:"true" default:"sheets"`
	ClixBucketName   string `split_words:"true" default:"clix"`
	PartsBucketName  string `split_words:"true" default:"parts"`
	PartsLockerKey   string `split_words:"true" default:"parts.lock"`
	TracksBucketName string `split_words:"true" default:"tracks"`
	MemberUser       string `split_words:"true" default:"admin"`
	MemberPass       string `split_words:"true" default:"admin"`
	PrepRepToken     string `split_words:"true" default:"admin"`
	AdminToken       string `split_words:"true" default:"admin"`
}

type FileBucket interface {
	PutFile(file *storage.File) bool
	DownloadURL(name string) (string, error)
}

type Storage struct {
	parts.Parts
	Sheets FileBucket
	Clix   FileBucket
	Tracks FileBucket
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
	navBar := NavBar{MemberUser: config.MemberUser}
	members := BasicAuth{config.MemberUser: config.MemberPass}
	//prepRep := TokenAuth{config.PrepRepToken, config.AdminToken}
	prepRep := PassThrough{}
	admin := TokenAuth{config.AdminToken}

	mux := http.NewServeMux()

	mux.Handle("/auth",
		prepRep.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	mux.Handle("/parts", members.Authenticate(PartsHandler{NavBar: navBar, Storage: database}))
	mux.Handle("/parts/", http.RedirectHandler("/parts", http.StatusMovedPermanently))

	downloadHandler := DownloadHandler{
		config.SheetsBucketName: database.Sheets.DownloadURL,
		config.ClixBucketName:   database.Clix.DownloadURL,
		config.TracksBucketName: database.Tracks.DownloadURL,
	}
	mux.Handle("/download", members.Authenticate(downloadHandler))

	// Uploads
	uploadHandler := UploadHandler{database}
	mux.Handle("/upload", prepRep.Authenticate(uploadHandler))

	mux.Handle("/login", members.Authenticate(http.RedirectHandler("/", http.StatusTemporaryRedirect)))

	mux.Handle("/version", http.HandlerFunc(Version))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			IndexHandler{NavBar: navBar}.ServeHTTP(w, r)
		} else {
			http.FileServer(http.Dir("public")).ServeHTTP(w, r)
		}
	}))

	return &http.Server{
		Addr:     config.ListenAddress,
		Handler:  mux,
		ErrorLog: log.StdLogger(),
	}
}
