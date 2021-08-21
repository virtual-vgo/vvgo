package server

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/api"
	"github.com/virtual-vgo/vvgo/pkg/server/api/arrangements"
	"github.com/virtual-vgo/vvgo/pkg/server/api/slash_command"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"github.com/virtual-vgo/vvgo/pkg/server/views"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
)

func authorize(role models.Role) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		identity := login.IdentityFromContext(r.Context())
		fmt.Println(identity)
		if !identity.HasRole(role) {
			helpers.Unauthorized(w)
		}
	}
}

func Routes() http.Handler {
	mux := RBACMux{ServeMux: http.NewServeMux()}

	// authorize
	for _, role := range []models.Role{models.RoleVVGOMember, models.RoleVVGOTeams, models.RoleVVGOLeader} {
		mux.HandleFunc("/authorize/"+role.String(), authorize(role), models.RoleAnonymous)
	}

	// debug
	mux.HandleFunc("/debug/pprof/", pprof.Index, models.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline, models.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile, models.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol, models.RoleVVGOTeams)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace, models.RoleVVGOTeams)

	// login
	mux.HandleFunc("/login/password", login.Password, models.RoleAnonymous)
	mux.HandleFunc("/login/discord", login.Discord, models.RoleAnonymous)
	mux.HandleFunc("/login/redirect", login.Redirect, models.RoleAnonymous)
	mux.HandleFunc("/login/success", views.LoginSuccess, models.RoleAnonymous)
	mux.HandleFunc("/login", views.Login, models.RoleAnonymous)
	mux.HandleFunc("/logout", login.Logout, models.RoleAnonymous)

	// api endpoints
	mux.HandleFunc("/api/v1/session", api.Session, models.RoleVVGOLeader)
	mux.HandleFunc("/api/v1/parts", api.Parts, models.RoleVVGOMember)
	mux.HandleFunc("/api/v1/projects", api.Projects, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/leaders", api.Leaders, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/roles", api.Roles, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/arrangements/ballot", arrangements.Ballot, models.RoleVVGOLeader)
	mux.HandleFunc("/api/v1/slash_commands", slash_command.Handle, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/slack_commands/list", slash_command.List, models.RoleVVGOTeams)
	mux.HandleFunc("/api/v1/slack_commands/update", slash_command.Update, models.RoleVVGOTeams)
	mux.HandleFunc("/api/v1/aboutme", api.Aboutme, models.RoleVVGOLeader)
	mux.HandleFunc("/api/v1/version", api.Version, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/update_stats", api.SkywardSwordIntentHandler, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/download", api.Download, models.RoleVVGOMember)
	mux.HandleFunc("/download", api.Download, models.RoleVVGOMember)

	// parts browser
	mux.Handle("/browser/static/",
		http.StripPrefix("/browser/", http.FileServer(http.Dir("ui/build"))),
		models.RoleVVGOMember)
	mux.HandleFunc("/browser/",
		func(w http.ResponseWriter, r *http.Request) {
			file, _ := os.Open("ui/build/index.html")
			io.Copy(w, file)
		}, models.RoleVVGOMember)

	// views
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
