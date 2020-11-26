package api

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net/http"
	"net/http/pprof"
)

var logger = log.Logger()

var PublicFiles = "public"

type ServerConfig struct {
	ListenAddress         string          `split_words:"true" default:"0.0.0.0:8080"`
	MemberUser            string          `split_words:"true" default:"admin"`
	MemberPass            string          `split_words:"true" default:"admin"`
	DistroBucketName      string          `split_words:"true" default:"vvgo-distro"`
	RedisNamespace        string          `split_words:"true" default:"local"`
	PartsSpreadsheetID    string          `envconfig:"parts_spreadsheet_id"`
	PartsReadRange        string          `envconfig:"parts_read_range"`
	DiscordGuildID        discord.GuildID `envconfig:"discord_guild_id"`
	DiscordRoleVVGOMember string          `envconfig:"discord_role_vvgo_member"`
	DiscordRoleVVGOTeams  string
	DiscordRoleVVGOLeader string
	Login                 login.Config `envconfig:"login"`
}

type Server struct {
	config   ServerConfig
	database Database
	*http.Server
}

func NewServer(ctx context.Context, config ServerConfig) *Server {
	var newBucket = func(ctx context.Context, bucketName string) *storage.Bucket {
		bucket, err := storage.NewBucket(ctx, bucketName)
		if err != nil {
			logger.WithError(err).WithField("bucket_name", bucketName).Fatal("storage.NewBucket() failed")
		}
		return bucket
	}

	database := Database{
		Distro:   newBucket(ctx, config.DistroBucketName),
		Sessions: login.NewStore(config.RedisNamespace, config.Login),
	}

	mux := RBACMux{
		ServeMux: http.NewServeMux(),
		Sessions: database.Sessions,
	}

	mux.Handle("/login/password", PasswordLoginHandler{
		Sessions: database.Sessions,
		Logins: map[[2]string][]login.Role{
			{config.MemberUser, config.MemberPass}: {login.RoleVVGOMember},
		},
	}, login.RoleAnonymous)

	mux.Handle("/login/discord", DiscordLoginHandler{
		GuildID:          config.DiscordGuildID,
		RoleVVGOLeaderID: config.DiscordRoleVVGOLeader,
		RoleVVGOTeamsID:  config.DiscordRoleVVGOTeams,
		RoleVVGOMemberID: config.DiscordRoleVVGOMember,
		Sessions:         database.Sessions,
	}, login.RoleAnonymous)

	mux.Handle("/login/success", LoginSuccessView{}, login.RoleAnonymous)

	mux.Handle("/login", LoginView{
		Sessions: database.Sessions,
	}, login.RoleAnonymous)

	mux.Handle("/logout", LogoutHandler{
		Sessions: database.Sessions,
	}, login.RoleAnonymous)

	mux.Handle("/roles", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity := IdentityFromContext(r.Context())
		jsonEncode(w, &identity.Roles)
	}), login.RoleAnonymous)

	// debug endpoints from net/http/pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index, login.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline, login.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile, login.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol, login.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace, login.RoleVVGOTeams)

	mux.Handle("/parts", PartView{}, login.RoleVVGOMember)

	mux.Handle("/archive", http.RedirectHandler("/projects/", http.StatusFound), login.RoleAnonymous)
	mux.Handle("/projects", http.RedirectHandler("/projects/", http.StatusFound), login.RoleAnonymous)
	mux.Handle("/projects/", ProjectsView{}, login.RoleAnonymous)

	mux.Handle("/download", DownloadHandler{
		config.DistroBucketName: database.Distro.DownloadURL,
	}, login.RoleVVGOMember)

	mux.Handle("/credits-maker", CreditsMaker{}, login.RoleVVGOTeams)

	mux.Handle("/about", AboutView{
		SpreadSheetID: config.PartsSpreadsheetID,
	}, login.RoleAnonymous)

	mux.Handle("/version", http.HandlerFunc(Version), login.RoleAnonymous)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			IndexView{}.ServeHTTP(w, r)
		} else {
			http.FileServer(http.Dir("public")).ServeHTTP(w, r)
		}
	}, login.RoleAnonymous)

	return &Server{
		config:   config,
		database: database,
		Server: &http.Server{
			Addr:     config.ListenAddress,
			Handler:  http_wrappers.Handler(&mux),
			ErrorLog: log.StdLogger(),
		},
	}
}
