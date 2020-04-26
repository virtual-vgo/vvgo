package api

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
	"net/http/pprof"
)

var logger = log.Logger()

var PublicFiles = "public"

type ServerConfig struct {
	ListenAddress         string `split_words:"true" default:"0.0.0.0:8080"`
	MaxContentLength      int64  `split_words:"true" default:"10000000"`
	MemberUser            string `split_words:"true" default:"admin"`
	MemberPass            string `split_words:"true" default:"admin"`
	PrepRepToken          string `split_words:"true" default:"admin"`
	AdminToken            string `split_words:"true" default:"admin"`
	DiscordLoginUrl       string `split_words:"true" default:"#"`
	DiscordGuildID        string `envconfig:"discord_guild_id"`
	DiscordRoleVVGOMember string `envconfig:"discord_role_vvgo_member"`
}

type StorageConfig struct {
	SheetsBucketName string `split_words:"true" default:"sheets"`
	ClixBucketName   string `split_words:"true" default:"clix"`
	PartsBucketName  string `split_words:"true" default:"parts"`
	PartsLockerKey   string `split_words:"true" default:"parts.lock"`
	TracksBucketName string `split_words:"true" default:"tracks"`
}

type Storage struct {
	StorageConfig
	Sessions *sessions.Store
	Parts    *parts.Parts
	Sheets   *storage.Bucket
	Clix     *storage.Bucket
	Tracks   *storage.Bucket
}

func NewStorage(ctx context.Context, sessions *sessions.Store, warehouse *storage.Warehouse, config StorageConfig) *Storage {
	var newBucket = func(ctx context.Context, bucketName string) *storage.Bucket {
		bucket, err := warehouse.NewBucket(ctx, config.SheetsBucketName)
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
		Sessions: sessions,
		Parts: &parts.Parts{
			Cache:  partsCache,
			Locker: partsLocker,
		},
		Sheets: sheetsBucket,
		Clix:   clixBucket,
		Tracks: tracksBucket,
	}
}

func (x *Storage) Init(ctx context.Context) error {
	if err := x.Sessions.Init(ctx); err != nil {
		return err
	} else if err = x.Parts.Init(ctx); err != nil {
		return err
	} else {
		logger.Info("storage initialized")
		return nil
	}
}

func NewServer(config ServerConfig, database *Storage, discordClient *discord.Client) *http.Server {
	navBar := NavBar{MemberUser: config.MemberUser, Sessions: database.Sessions, DiscordLoginUrl: config.DiscordLoginUrl}
	rbacMux := NewRBACMux(database.Sessions)
	rbacMux.Handle("/version", http.HandlerFunc(Version), sessions.RoleAnonymous)

	// debug endpoints from net/http/pprof
	pprofMux := http.NewServeMux()
	pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
	pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	rbacMux.Handle("/debug/pprof/", pprofMux, sessions.RoleVVGODeveloper)

	// authentication handlers
	passwordLoginHandler := PasswordLoginHandler{
		Sessions: database.Sessions,
		Logins: []PasswordLogin{
			{
				User:  config.MemberUser,
				Pass:  config.MemberPass,
				Roles: []sessions.Role{sessions.RoleVVGOMember},
			},
		},
	}
	rbacMux.Handle("/auth/password", passwordLoginHandler, sessions.RoleAnonymous)

	discordLoginHandler := DiscordLoginHandler{
		GuildID:        discord.GuildID(config.DiscordGuildID),
		RoleVVGOMember: config.DiscordRoleVVGOMember,
		Sessions:       database.Sessions,
		Discord:        discordClient,
	}
	rbacMux.Handle("/auth/discord", discordLoginHandler, sessions.RoleAnonymous)

	logoutHandler := LogoutHandler{Sessions: database.Sessions}
	rbacMux.Handle("/logout", logoutHandler, sessions.RoleAnonymous)

	// Upload

	uploadHandler := UploadHandler{database}
	rbacMux.Handle("/upload", uploadHandler, sessions.RoleVVGOUploader)

	// Download

	downloadHandler := DownloadHandler{
		database.SheetsBucketName: database.Sheets.DownloadURL,
		database.ClixBucketName:   database.Clix.DownloadURL,
		database.TracksBucketName: database.Tracks.DownloadURL,
	}
	rbacMux.Handle("/download", downloadHandler, sessions.RoleVVGOMember)

	// Views
	partsView := PartView{NavBar: navBar, Storage: database}
	rbacMux.Handle("/parts", partsView, sessions.RoleVVGOMember)

	loginView := LoginView{NavBar: navBar, Sessions: database.Sessions}
	rbacMux.Handle("/login", loginView, sessions.RoleAnonymous)

	indexView := IndexView{NavBar: navBar}
	rbacMux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			indexView.ServeHTTP(w, r)
		} else {
			http.FileServer(http.Dir("public")).ServeHTTP(w, r)
		}
	}), sessions.RoleAnonymous)

	return &http.Server{
		Addr:     config.ListenAddress,
		Handler:  tracing.WrapHandler(rbacMux),
		ErrorLog: log.StdLogger(),
	}
}
