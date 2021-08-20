package api

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api/config"
	"github.com/virtual-vgo/vvgo/pkg/api/helpers"
	"github.com/virtual-vgo/vvgo/pkg/api/session"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
)

var PublicFiles = "public"

func Routes() http.Handler {
	mux := RBACMux{ServeMux: http.NewServeMux()}

	mux.Handle("/login/password", PasswordLoginHandler{}, login.RoleAnonymous)
	mux.Handle("/login/discord", DiscordLoginHandler{}, login.RoleAnonymous)
	mux.Handle("/login/success", LoginSuccessView{}, login.RoleAnonymous)
	mux.Handle("/login/redirect", LoginRedirect{}, login.RoleAnonymous)
	mux.Handle("/login", LoginView{}, login.RoleAnonymous)
	mux.Handle("/logout", LogoutHandler{}, login.RoleAnonymous)

	for _, role := range []login.Role{login.RoleVVGOMember, login.RoleVVGOTeams, login.RoleVVGOLeader} {
		func(role login.Role) {
			mux.Handle("/authorize/"+role.String(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				identity := IdentityFromContext(r.Context())
				fmt.Println(identity)
				if !identity.HasRole(role) {
					helpers.Unauthorized(w)
				}
			}), login.RoleAnonymous)
		}(role)
	}

	// debug endpoints from net/http/pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index, login.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline, login.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile, login.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol, login.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace, login.RoleVVGOTeams)

	//mux.HandleFunc("/api/v1/config", ConfigApi, login.RoleReadConfig)
	mux.HandleFunc("/api/v1/session", session.Handler, login.RoleVVGOLeader)
	mux.HandleFunc("/api/v1/config", config.Handler, login.RoleReadConfig)
	mux.HandleFunc("/api/v1/parts", PartsApi, login.RoleVVGOMember)
	mux.HandleFunc("/api/v1/projects", ProjectsApi, login.RoleAnonymous)
	mux.HandleFunc("/api/v1/leaders", LeadersApi, login.RoleAnonymous)
	mux.HandleFunc("/api/v1/roles", RolesApi, login.RoleAnonymous)
	mux.HandleFunc("/api/v1/arrangements/ballot", ArrangementsBallotApi, login.RoleVVGOLeader)
	mux.HandleFunc("/api/v1/slash_commands", HandleSlashCommand, login.RoleAnonymous)
	mux.HandleFunc("/api/v1/update_stats", SkywardSwordIntentHandler, login.RoleAnonymous)
	mux.HandleFunc("/api/v1/aboutme", AboutMeApi, login.RoleVVGOLeader)

	mux.Handle("/browser/static/",
		http.StripPrefix("/browser/", http.FileServer(http.Dir("ui/build"))),
		login.RoleVVGOMember)
	mux.HandleFunc("/browser/",
		func(w http.ResponseWriter, r *http.Request) {
			file, _ := os.Open("ui/build/index.html")
			io.Copy(w, file)
		}, login.RoleVVGOMember)

	mux.HandleFunc("/slash_commands", ViewSlashCommands, login.RoleVVGOTeams)
	mux.HandleFunc("/slash_commands/create", CreateSlashCommands, login.RoleVVGOTeams)

	mux.HandleFunc("/voting", VotingView, login.RoleVVGOLeader)
	mux.HandleFunc("/voting/results", VotingResultsView, login.RoleVVGOLeader)
	mux.HandleFunc("/parts", PartsView, login.RoleVVGOMember)
	mux.HandleFunc("/projects", ProjectsView, login.RoleAnonymous)
	mux.HandleFunc("/download", DownloadHandler, login.RoleVVGOMember)
	mux.HandleFunc("/credits-maker", CreditsMaker, login.RoleVVGOTeams)
	mux.HandleFunc("/about", AboutView, login.RoleAnonymous)
	mux.HandleFunc("/version", Version, login.RoleAnonymous)
	mux.HandleFunc("/contact_us", ContactUs, login.RoleAnonymous)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			IndexView(w, r)
		} else {
			http.FileServer(http.Dir(PublicFiles)).ServeHTTP(w, r)
		}
	}, login.RoleAnonymous)
	return &mux
}
