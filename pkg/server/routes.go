package server

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/api"
	"github.com/virtual-vgo/vvgo/pkg/server/api/arrangements"
	"github.com/virtual-vgo/vvgo/pkg/server/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/server/api/devel"
	"github.com/virtual-vgo/vvgo/pkg/server/api/mixtape"
	"github.com/virtual-vgo/vvgo/pkg/server/api/slash_command"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"io"
	"net/http"
	"net/http/pprof"
	"os"
)

func Routes() http.Handler {
	mux := RBACMux{ServeMux: http.NewServeMux()}

	// authorize
	for _, role := range []models.Role{models.RoleVVGOVerifiedMember, models.RoleVVGOProductionTeam, models.RoleVVGOExecutiveDirector} {
		mux.HandleFunc("/authorize/"+role.String(), authorize(role), models.RoleAnonymous)
	}

	// debug
	mux.HandleFunc("/debug/pprof/", pprof.Index, models.RoleVVGOProductionTeam)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline, models.RoleVVGOProductionTeam)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile, models.RoleVVGOProductionTeam)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol, models.RoleVVGOProductionTeam)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace, models.RoleVVGOProductionTeam)

	// api endpoints
	mux.HandleApiFunc("/api/v1/arrangements/ballot", arrangements.Ballot, models.RoleVVGOExecutiveDirector)
	mux.HandleApiFunc("/api/v1/auth/discord", auth.Discord, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/auth/logout", auth.Logout, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/auth/oauth_redirect", auth.OAuthRedirect, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/credits", api.Credits, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/credits/pasta", api.CreditsPasta, models.RoleVVGOProductionTeam)
	mux.HandleApiFunc("/api/v1/credits/table", api.CreditsTable, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/dataset", api.Dataset, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/guild_members", api.GuildMembers, models.RoleVVGOExecutiveDirector)
	mux.HandleApiFunc("/api/v1/auth/password", auth.Password, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/me", api.Me, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/mixtape", mixtape.ProjectsHandler, models.RoleVVGOVerifiedMember)
	mux.HandleApiFunc("/api/v1/parts", api.Parts, models.RoleVVGOVerifiedMember)
	mux.HandleApiFunc("/api/v1/projects", api.Projects, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/sessions", api.Sessions, models.RoleVVGOExecutiveDirector)
	mux.HandleFunc("/api/v1/slash_commands", slash_command.Handle, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/slack_commands/list", slash_command.List, models.RoleVVGOProductionTeam)
	mux.HandleFunc("/api/v1/slack_commands/update", slash_command.Update, models.RoleVVGOProductionTeam)
	mux.HandleApiFunc("/api/v1/spreadsheet", api.Spreadsheet, models.RoleWriteSpreadsheet)
	mux.HandleApiFunc("/api/v1/version", api.Version, models.RoleAnonymous)
	mux.HandleFunc("/download", api.Download, models.RoleVVGOVerifiedMember)

	if config.Config.Development {
		mux.HandleFunc("/api/v1/devel/fetch_spreadsheets", devel.FetchSpreadsheets, models.RoleVVGOProductionTeam)
	}

	// views
	mux.HandleFunc("/parts/", serveUI, models.RoleAnonymous)
	mux.HandleFunc("/projects/", serveUI, models.RoleAnonymous)
	mux.HandleFunc("/credits-maker/", serveUI, models.RoleAnonymous)
	mux.HandleFunc("/about/", serveUI, models.RoleAnonymous)
	mux.HandleFunc("/contact/", serveUI, models.RoleAnonymous)
	mux.HandleFunc("/sessions/", serveUI, models.RoleAnonymous)
	mux.HandleFunc("/mixtape/", serveUI, models.RoleAnonymous)
	mux.HandleFunc("/login/", serveUI, models.RoleAnonymous)
	mux.HandleFunc("/logout/", serveUI, models.RoleAnonymous)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			serveUI(w, r)
		} else {
			http.FileServer(http.Dir("public")).ServeHTTP(w, r)
		}
	}, models.RoleAnonymous)
	return &mux
}

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

func serveUI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	file, err := os.Open("public/index.html")
	if err != nil {
		logger.OpenFileFailure(ctx, err)
		http_helpers.WriteInternalServerError(ctx, w)
		return
	}
	if _, err := io.Copy(w, file); err != nil {
		logger.MethodFailure(ctx, "io.Copy", err)
	}
}
