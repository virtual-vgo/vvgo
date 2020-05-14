package api

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"net/http/pprof"
)

var logger = log.Logger()

var PublicFiles = "public"

type ServerConfig struct {
	ListenAddress    string `split_words:"true" default:"0.0.0.0:8080"`
	MaxContentLength int64  `split_words:"true" default:"10000000"`
	MemberUser       string `split_words:"true" default:"admin"`
	MemberPass       string `split_words:"true" default:"admin"`
	PrepRepToken     string `split_words:"true" default:"admin"`
	AdminToken       string `split_words:"true" default:"admin"`
	SheetsBucketName string `split_words:"true" default:"sheets"`
	ClixBucketName   string `split_words:"true" default:"clix"`
	TracksBucketName string `split_words:"true" default:"tracks"`
	RedisNamespace   string `split_words:"true" default:"local"`
}

type Database struct {
	ServerConfig
	Parts  *parts.RedisParts
	Sheets *storage.Bucket
	Clix   *storage.Bucket
	Tracks *storage.Bucket
}

func NewServer(ctx context.Context, config ServerConfig) *http.Server {
	var newBucket = func(ctx context.Context, bucketName string) *storage.Bucket {
		bucket, err := storage.NewBucket(ctx, bucketName)
		if err != nil {
			logger.WithError(err).WithField("bucket_name", bucketName).Fatal("warehouse.NewBucket() failed")
		}
		return bucket
	}

	database := Database{
		ServerConfig: config,
		Sheets:       newBucket(ctx, config.SheetsBucketName),
		Clix:         newBucket(ctx, config.ClixBucketName),
		Tracks:       newBucket(ctx, config.TracksBucketName),
		Parts:        parts.NewParts(config.RedisNamespace + ":parts"),
	}

	navBar := NavBar{MemberUser: config.MemberUser}
	members := BasicAuth{config.MemberUser: config.MemberPass}
	prepRep := TokenAuth{config.PrepRepToken, config.AdminToken}
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

	partsHandler := members.Authenticate(&PartView{NavBar: navBar, Database: &database})
	mux.Handle("/parts", partsHandler)
	mux.Handle("/parts/", http.RedirectHandler("/parts", http.StatusMovedPermanently))

	downloadHandler := members.Authenticate(&DownloadHandler{
		config.SheetsBucketName: database.Sheets.DownloadURL,
		config.ClixBucketName:   database.Clix.DownloadURL,
		config.TracksBucketName: database.Tracks.DownloadURL,
	})
	mux.Handle("/download", members.Authenticate(downloadHandler))

	// Uploads
	uploadHandler := prepRep.Authenticate(&UploadHandler{&database})
	mux.Handle("/upload", uploadHandler)

	loginHandler := members.Authenticate(http.RedirectHandler("/", http.StatusTemporaryRedirect))
	mux.Handle("/login", loginHandler)

	mux.Handle("/version", http.HandlerFunc(Version))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			IndexView{NavBar: navBar}.ServeHTTP(w, r)
		} else {
			http.FileServer(http.Dir("public")).ServeHTTP(w, r)
		}
	}))

	return &http.Server{
		Addr:     config.ListenAddress,
		Handler:  tracing.WrapHandler(mux),
		ErrorLog: log.StdLogger(),
	}
}
