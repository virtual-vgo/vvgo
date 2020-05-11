package api

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/login"
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
	UploaderToken    string `split_words:"true" default:"admin"`
	DeveloperToken   string `split_words:"true" default:"admin"`
}

type StorageConfig struct {
	SheetsBucketName string       `split_words:"true" default:"sheets"`
	ClixBucketName   string       `split_words:"true" default:"clix"`
	TracksBucketName string       `split_words:"true" default:"tracks"`
	RedisNamespace   string       `split_words:"true" default:"local"`
	SessionsConfig   login.Config `envconfig:"sessions"`
}

type Storage struct {
	StorageConfig
	Parts    *parts.RedisParts
	Sheets   *storage.Bucket
	Clix     *storage.Bucket
	Tracks   *storage.Bucket
	Sessions *login.Store
}

func NewStorage(ctx context.Context, warehouse *storage.Warehouse, config StorageConfig) *Storage {
	var newBucket = func(ctx context.Context, bucketName string) *storage.Bucket {
		bucket, err := warehouse.NewBucket(ctx, bucketName)
		if err != nil {
			logger.WithError(err).WithField("bucket_name", bucketName).Fatal("warehouse.NewBucket() failed")
		}
		return bucket
	}

	db := Storage{
		StorageConfig: config,
		Sheets:        newBucket(ctx, config.SheetsBucketName),
		Clix:          newBucket(ctx, config.ClixBucketName),
		Tracks:        newBucket(ctx, config.TracksBucketName),
		Parts:         parts.NewParts(config.RedisNamespace),
		Sessions:      login.NewStore(config.RedisNamespace, config.SessionsConfig),
	}
	return &db
}

func NewServer(config ServerConfig, database *Storage) *http.Server {
	rbacMux := RBACMux{
		Basic: map[[2]string][]login.Role{
			{config.MemberUser, config.MemberPass}: {login.RoleVVGOMember},
		},
		Bearer: map[string][]login.Role{
			config.UploaderToken:  {login.RoleVVGOUploader, login.RoleVVGOMember},
			config.DeveloperToken: {login.RoleVVGODeveloper, login.RoleVVGOMember},
		},
		Sessions: database.Sessions,
		ServeMux: http.NewServeMux(),
	}

	rbacMux.Handle("/login", LoginView{
		Sessions: database.Sessions,
	}, login.RoleAnonymous)

	rbacMux.Handle("/login/password", PasswordLoginHandler{
		Sessions: database.Sessions,
		Logins: map[[2]string][]login.Role{
			{config.MemberUser, config.MemberPass}: {login.RoleVVGOMember},
		},
	}, login.RoleAnonymous)

	rbacMux.Handle("/logout", LogoutHandler{Sessions: database.Sessions}, login.RoleAnonymous)

	rbacMux.Handle("/auth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("authenticated"))
	}), login.RoleVVGOMember)

	// debug endpoints from net/http/pprof
	rbacMux.HandleFunc("/debug/pprof/", pprof.Index, login.RoleVVGODeveloper)
	rbacMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline, login.RoleVVGODeveloper)
	rbacMux.HandleFunc("/debug/pprof/profile", pprof.Profile, login.RoleVVGODeveloper)
	rbacMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol, login.RoleVVGODeveloper)
	rbacMux.HandleFunc("/debug/pprof/trace", pprof.Trace, login.RoleVVGODeveloper)

	rbacMux.Handle("/parts", PartView{Storage: database}, login.RoleVVGOMember)

	rbacMux.Handle("/download", DownloadHandler{
		database.SheetsBucketName: database.Sheets.DownloadURL,
		database.ClixBucketName:   database.Clix.DownloadURL,
		database.TracksBucketName: database.Tracks.DownloadURL,
	}, login.RoleVVGOMember)

	// Uploads
	rbacMux.Handle("/upload", UploadHandler{
		Storage: database,
	}, login.RoleVVGOUploader)

	rbacMux.Handle("/version", http.HandlerFunc(Version), login.RoleAnonymous)
	rbacMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			IndexView{}.ServeHTTP(w, r)
		} else {
			http.FileServer(http.Dir("public")).ServeHTTP(w, r)
		}
	}, login.RoleAnonymous)

	return &http.Server{
		Addr:     config.ListenAddress,
		Handler:  tracing.WrapHandler(&rbacMux),
		ErrorLog: log.StdLogger(),
	}
}
