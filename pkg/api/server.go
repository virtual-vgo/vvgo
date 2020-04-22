package api

import (
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"net/http/httputil"
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
	SessionsKey      string `split_words:"true" default:"sessions"`
	DiscordOAuthHandlerConfig `envconfig:"discord"`
}

type Storage struct {
	parts.Parts
	Sheets *storage.Bucket
	Clix   *storage.Bucket
	Tracks *storage.Bucket
	ServerConfig
}

func NewStorage(ctx context.Context, config ServerConfig) *Storage {
	var newBucket = func(ctx context.Context, bucketName string) *storage.Bucket {
		bucket, err := storage.NewBucket(ctx, config.SheetsBucketName)
		if err != nil {
			logger.WithError(err).WithField("bucket_name", config.SheetsBucketName).Fatal("warehouse.NewBucket() failed")
		}
		return bucket
	}

	sheetsBucket := newBucket(ctx, config.SheetsBucketName)
	clixBucket := newBucket(ctx, config.ClixBucketName)
	tracksBucket := newBucket(ctx, config.TracksBucketName)
	partsCache := storage.NewCache(storage.CacheOpts{Bucket: newBucket(ctx, config.PartsBucketName)})
	partsLocker := locker.NewLocker(locker.Opts{RedisKey: config.PartsLockerKey})

	return &Storage{
		Parts: parts.Parts{
			Cache:  partsCache,
			Locker: partsLocker,
		},
		Sheets:       sheetsBucket,
		Clix:         clixBucket,
		Tracks:       tracksBucket,
		ServerConfig: config,
	}
}

func NewServer(config ServerConfig, database *Storage) *http.Server {
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

	partsHandler := members.Authenticate(&PartsHandler{NavBar: navBar, Storage: database})
	mux.Handle("/parts", partsHandler)
	mux.Handle("/parts/", http.RedirectHandler("/parts", http.StatusMovedPermanently))

	downloadHandler := members.Authenticate(&DownloadHandler{
		config.SheetsBucketName: database.Sheets.DownloadURL,
		config.ClixBucketName:   database.Clix.DownloadURL,
		config.TracksBucketName: database.Tracks.DownloadURL,
	})
	mux.Handle("/download", members.Authenticate(downloadHandler))

	// Uploads
	uploadHandler := prepRep.Authenticate(&UploadHandler{database})
	mux.Handle("/upload", uploadHandler)

	loginHandler := &LoginHandler{
		NavBar:   navBar,
		Sessions: sessions.NewStore(sessions.StoreOpts{LockerName: config.SessionsKey}),
	}
	mux.Handle("/login", loginHandler)

	mux.Handle("/oauth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := httputil.DumpRequest(r, true)
		fmt.Println(string(body))


	}))

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
		Handler:  tracing.WrapHandler(mux),
		ErrorLog: log.StdLogger(),
	}
}
