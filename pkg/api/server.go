package api

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/discord"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"net/http"
	"net/http/pprof"
)

var logger = log.Logger()

var PublicFiles = "public"

type ServerConfig struct {
	ListenAddress         string          `split_words:"true" default:"0.0.0.0:8080"`
	MemberUser            string          `split_words:"true" default:"admin"`
	MemberPass            string          `split_words:"true" default:"admin"`
	RedisNamespace        string          `split_words:"true" default:"local"`
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

func NewServer(ctx context.Context, serverConfig ServerConfig) *Server {
	template := Template{}

	mux := RBACMux{ServeMux: http.NewServeMux()}

	mux.Handle("/login/password", PasswordLoginHandler{
		Logins: map[[2]string][]login.Role{
			{serverConfig.MemberUser, serverConfig.MemberPass}: {login.RoleVVGOMember},
		},
	}, login.RoleAnonymous)

	mux.Handle("/login/discord", DiscordLoginHandler{
		GuildID:          serverConfig.DiscordGuildID,
		RoleVVGOLeaderID: serverConfig.DiscordRoleVVGOLeader,
		RoleVVGOTeamsID:  serverConfig.DiscordRoleVVGOTeams,
		RoleVVGOMemberID: serverConfig.DiscordRoleVVGOMember,
	}, login.RoleAnonymous)

	mux.Handle("/login/success", LoginSuccessView{template}, login.RoleAnonymous)
	mux.Handle("/login/redirect", LoginRedirect{}, login.RoleAnonymous)

	mux.Handle("/login", LoginView{Template: template}, login.RoleAnonymous)

	mux.Handle("/logout", LogoutHandler{}, login.RoleAnonymous)

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

	mux.Handle("/parts", PartView{template}, login.RoleVVGOMember)

	mux.Handle("/archive", http.RedirectHandler("/projects/", http.StatusFound), login.RoleAnonymous)
	mux.Handle("/projects", http.RedirectHandler("/projects/", http.StatusFound), login.RoleAnonymous)
	mux.Handle("/projects/", ProjectsView{template}, login.RoleAnonymous)

	mux.Handle("/download", DownloadHandler{}, login.RoleVVGOMember)

	mux.Handle("/credits-maker", CreditsMaker{template}, login.RoleVVGOTeams)

	mux.Handle("/about", AboutView{template}, login.RoleAnonymous)

	mux.Handle("/version", http.HandlerFunc(Version), login.RoleAnonymous)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			IndexView{template}.ServeHTTP(w, r)
		} else {
			http.FileServer(http.Dir("public")).ServeHTTP(w, r)
		}
	}, login.RoleAnonymous)

	return &Server{
		config:   serverConfig,
		Server: &http.Server{
			Addr:     serverConfig.ListenAddress,
			Handler:  http_wrappers.Handler(&mux),
			ErrorLog: log.StdLogger(),
		},
	}
}
