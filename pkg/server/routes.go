package server

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/api"
	"github.com/virtual-vgo/vvgo/pkg/server/api/aboutme"
	"github.com/virtual-vgo/vvgo/pkg/server/api/arrangements"
	"github.com/virtual-vgo/vvgo/pkg/server/api/download"
	"github.com/virtual-vgo/vvgo/pkg/server/api/leaders"
	"github.com/virtual-vgo/vvgo/pkg/server/api/parts"
	"github.com/virtual-vgo/vvgo/pkg/server/api/projects"
	"github.com/virtual-vgo/vvgo/pkg/server/api/roles"
	"github.com/virtual-vgo/vvgo/pkg/server/api/session"
	"github.com/virtual-vgo/vvgo/pkg/server/api/slash_command"
	"github.com/virtual-vgo/vvgo/pkg/server/api/version"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"github.com/virtual-vgo/vvgo/pkg/server/views"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
)

func Routes() http.Handler {
	mux := RBACMux{ServeMux: http.NewServeMux()}

	mux.Handle("/login/password", login.PasswordLoginHandler{}, models.RoleAnonymous)
	mux.Handle("/login/discord", login.DiscordLoginHandler{}, models.RoleAnonymous)
	mux.Handle("/login/success", views.LoginSuccessView{}, models.RoleAnonymous)
	mux.Handle("/login/redirect", login.Redirect{}, models.RoleAnonymous)
	mux.Handle("/login", views.LoginView{}, models.RoleAnonymous)
	mux.Handle("/logout", login.LogoutHandler{}, models.RoleAnonymous)

	for _, role := range []models.Role{models.RoleVVGOMember, models.RoleVVGOTeams, models.RoleVVGOLeader} {
		func(role models.Role) {
			mux.Handle("/authorize/"+role.String(), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				identity := login.IdentityFromContext(r.Context())
				fmt.Println(identity)
				if !identity.HasRole(role) {
					helpers.Unauthorized(w)
				}
			}), models.RoleAnonymous)
		}(role)
	}

	// debug endpoints from net/http/pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index, models.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline, models.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile, models.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol, models.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace, models.RoleVVGOTeams)

	// api endpoints
	mux.HandleFunc("/api/v1/session", session.Handle, models.RoleVVGOLeader)
	mux.HandleFunc("/api/v1/parts", parts.Handle, models.RoleVVGOMember)
	mux.HandleFunc("/api/v1/projects", projects.Handle, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/leaders", leaders.Handle, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/roles", roles.Handle, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/arrangements/ballot", arrangements.Ballot, models.RoleVVGOLeader)
	mux.HandleFunc("/api/v1/slash_commands", slash_command.Handle, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/slack_commands/list", slash_command.List, models.RoleVVGOTeams)
	mux.HandleFunc("/api/v1/slack_commands/update", slash_command.Update, models.RoleVVGOTeams)
	mux.HandleFunc("/api/v1/aboutme", aboutme.Handle, models.RoleVVGOLeader)
	mux.HandleFunc("/api/v1/version", version.Handle, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/update_stats", api.SkywardSwordIntentHandler, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/download", download.Handler, models.RoleVVGOMember)
	mux.HandleFunc("/download", download.Handler, models.RoleVVGOMember)

	mux.Handle("/browser/static/",
		http.StripPrefix("/browser/", http.FileServer(http.Dir("ui/build"))),
		models.RoleVVGOMember)
	mux.HandleFunc("/browser/",
		func(w http.ResponseWriter, r *http.Request) {
			file, _ := os.Open("ui/build/index.html")
			io.Copy(w, file)
		}, models.RoleVVGOMember)

	mux.HandleFunc("/voting", views.Voting, models.RoleVVGOLeader)
	mux.HandleFunc("/voting/results", views.VotingResults, models.RoleVVGOLeader)
	mux.HandleFunc("/parts", views.Parts, models.RoleVVGOMember)
	mux.HandleFunc("/projects", views.Projects, models.RoleAnonymous)
	mux.HandleFunc("/credits-maker", views.CreditsMaker, models.RoleVVGOTeams)
	mux.HandleFunc("/about", views.About, models.RoleAnonymous)
	mux.HandleFunc("/contact_us", views.ContactUs, models.RoleAnonymous)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			views.Index(w, r)
		} else {
			views.ServePublicFile(w, r)
		}
	}, models.RoleAnonymous)
	return &mux
}
