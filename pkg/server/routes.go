package server

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/api"
	"github.com/virtual-vgo/vvgo/pkg/server/api/arrangements"
	"github.com/virtual-vgo/vvgo/pkg/server/api/devel"
	"github.com/virtual-vgo/vvgo/pkg/server/api/mixtape"
	"github.com/virtual-vgo/vvgo/pkg/server/api/slash_command"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"github.com/virtual-vgo/vvgo/pkg/server/views"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
)

func authorize(role models.Role) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		identity := login.IdentityFromContext(r.Context())
		fmt.Println(identity)
		if !identity.HasRole(role) {
			http_helpers.WriteUnauthorizedError(ctx, w)
		}
	}
}

func Routes() http.Handler {
	mux := RBACMux{ServeMux: http.NewServeMux()}

	// authorize
	for _, role := range []models.Role{models.RoleVVGOMember, models.RoleVVGOTeams, models.RoleVVGOExecutiveDirector} {
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
	mux.HandleFunc("/api/v1/aboutme", api.Aboutme, models.RoleVVGOExecutiveDirector)
	mux.HandleFunc("/api/v1/arrangements/ballot", arrangements.Ballot, models.RoleVVGOExecutiveDirector)
	mux.HandleFunc("/api/v1/credits", api.Credits, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/download", api.Download, models.RoleVVGOMember)
	mux.HandleFunc("/api/v1/dataset", api.Dataset, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/guild_members", api.GuildMembers, models.RoleVVGOExecutiveDirector)
	mux.HandleFunc("/api/v1/me", api.Me, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/mixtape", mixtape.Handler, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/parts", api.Parts, models.RoleVVGOMember)
	mux.HandleFunc("/api/v1/projects", api.Projects, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/roles", api.Roles, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/sessions", api.Sessions, models.RoleVVGOExecutiveDirector)
	mux.HandleFunc("/api/v1/slash_commands", slash_command.Handle, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/slack_commands/list", slash_command.List, models.RoleVVGOTeams)
	mux.HandleFunc("/api/v1/slack_commands/update", slash_command.Update, models.RoleVVGOTeams)
	mux.HandleFunc("/api/v1/spreadsheet", api.Spreadsheet, models.RoleWriteSpreadsheet)
	mux.HandleFunc("/api/v1/version", api.Version, models.RoleAnonymous)
	mux.HandleFunc("/download", api.Download, models.RoleVVGOMember)

	if config.Config.Development {
		mux.HandleFunc("/api/v1/devel/fetch_spreadsheets", devel.FetchSpreadsheets, models.RoleVVGOTeams)
	}

	// parts browser
	mux.Handle("/browser/static/",
		http.StripPrefix("/browser/", http.FileServer(http.Dir("parts_browser/build"))),
		models.RoleVVGOMember)
	mux.HandleFunc("/browser/",
		func(w http.ResponseWriter, r *http.Request) {
			file, _ := os.Open("parts_browser/build/index.html")
			io.Copy(w, file)
		}, models.RoleVVGOMember)

	// views
	mux.HandleFunc("/voting", views.Voting, models.RoleVVGOExecutiveDirector)
	mux.HandleFunc("/voting/results", views.VotingResults, models.RoleVVGOExecutiveDirector)
	mux.HandleFunc("/parts", views.Parts, models.RoleVVGOMember)
	mux.HandleFunc("/projects", views.Projects, models.RoleAnonymous)
	mux.HandleFunc("/credits-maker", views.CreditsMaker, models.RoleVVGOTeams)
	mux.HandleFunc("/about", views.About, models.RoleAnonymous)
	mux.HandleFunc("/contact_us", views.ContactUs, models.RoleAnonymous)
	mux.HandleFunc("/feature", views.ServeTemplate("feature.gohtml"), models.RoleAnonymous)
	mux.HandleFunc("/sessions", views.Sessions, models.RoleVVGOMember)
	mux.HandleFunc("/mixtape", views.Mixtape, models.RoleVVGOMember)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			views.Index(w, r)
		} else {
			views.ServePublicFile(w, r)
		}
	}, models.RoleAnonymous)
	return &mux
}
