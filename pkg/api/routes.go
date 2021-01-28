package api

import (
	"fmt"
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
					unauthorized(w)
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

	mux.Handle("/api/v1/parts", PartsAPI{}, login.RoleVVGOMember)
	mux.Handle("/api/v1/projects", ProjectsAPI{}, login.RoleAnonymous)
	mux.Handle("/api/v1/leaders", LeadersAPI{}, login.RoleAnonymous)
	mux.Handle("/api/v1/roles", RolesAPI{}, login.RoleAnonymous)
	mux.Handle("/api/v1/arrangements/ballot", ArrangementsBallotAPI, login.RoleVVGOLeader)

	mux.Handle("/voting", VotingView, login.RoleVVGOLeader)
	mux.Handle("/voting/results", VotingResultsView{}, login.RoleVVGOLeader)

	mux.Handle("/browser/static/",
		http.StripPrefix("/browser/", http.FileServer(http.Dir("ui/build"))),
		login.RoleVVGOMember)
	mux.HandleFunc("/browser/",
		func(w http.ResponseWriter, r *http.Request) {
			file, _ := os.Open("ui/build/index.html")
			io.Copy(w, file)
		}, login.RoleVVGOMember)

	mux.Handle("/parts", PartView{}, login.RoleVVGOMember)
	mux.Handle("/projects", ProjectsView{}, login.RoleAnonymous)
	mux.Handle("/download", DownloadHandler{}, login.RoleVVGOMember)
	mux.Handle("/credits-maker", CreditsMaker{}, login.RoleVVGOTeams)
	mux.Handle("/about", AboutView{}, login.RoleAnonymous)
	mux.Handle("/version", http.HandlerFunc(Version), login.RoleAnonymous)
	mux.Handle("/contact_us", ContactUs, login.RoleAnonymous)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			IndexView{}.ServeHTTP(w, r)
		} else {
			http.FileServer(http.Dir(PublicFiles)).ServeHTTP(w, r)
		}
	}, login.RoleAnonymous)
	return &mux
}
