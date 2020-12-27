package api

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"net/http"
	"net/http/pprof"
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
	mux.Handle("/projects", ProjectsView{}, login.RoleAnonymous)
	mux.Handle("/download", DownloadHandler{}, login.RoleVVGOMember)
	mux.Handle("/credits-maker", CreditsMaker{}, login.RoleVVGOTeams)
	mux.Handle("/about", AboutView{}, login.RoleAnonymous)
	mux.Handle("/version", http.HandlerFunc(Version), login.RoleAnonymous)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			IndexView{}.ServeHTTP(w, r)
		} else {
			http.FileServer(http.Dir(PublicFiles)).ServeHTTP(w, r)
		}
	}, login.RoleAnonymous)
	return &mux
}
