package server

import (
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/api"
	"github.com/virtual-vgo/vvgo/pkg/server/api/arrangements"
	"github.com/virtual-vgo/vvgo/pkg/server/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/server/api/devel"
	"github.com/virtual-vgo/vvgo/pkg/server/api/guild_members"
	"github.com/virtual-vgo/vvgo/pkg/server/api/mixtape"
	"github.com/virtual-vgo/vvgo/pkg/server/api/slash_command"
	"github.com/virtual-vgo/vvgo/pkg/server/api/traces"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
)

const PublicFiles = "public"

var ServeUI = http.FileServer(Filesystem("public"))

type Filesystem string

func (fs Filesystem) Open(name string) (http.File, error) {
	file, err := os.Open(path.Join(PublicFiles, "dist", name))
	if errors.Is(err, os.ErrNotExist) {
		return os.Open(path.Join(PublicFiles, "dist", "index.html"))
	}
	return file, err
}

func authorize(role models.Role) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		identity := login.IdentityFromContext(r.Context())
		fmt.Println(identity)
		if !identity.HasRole(role) {
			http_helpers.WriteAPIResponse(ctx, w, http_helpers.NewUnauthorizedError())
		}
	}
}

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
	mux.HandleApiFunc("/api/v1/download", api.Download, models.RoleDownload)
	mux.HandleApiFunc("/api/v1/guild_members/search", guild_members.HandleSearch, models.RoleVVGOVerifiedMember)
	mux.HandleApiFunc("/api/v1/guild_members/lookup", guild_members.HandleLookup, models.RoleVVGOVerifiedMember)
	//mux.HandleApiFunc("/api/v1/guild_members/list", guild_members.HandleList, models.RoleVVGOExecutiveDirector)
	mux.HandleApiFunc("/api/v1/auth/password", auth.Password, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/traces/spans", traces.HandleSpans, models.RoleVVGOProductionTeam)
	mux.HandleApiFunc("/api/v1/traces/waterfall", traces.HandleWaterfall, models.RoleVVGOExecutiveDirector)
	mux.HandleApiFunc("/api/v1/me", api.Me, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/mixtape/projects", mixtape.HandleProjects, models.RoleVVGOVerifiedMember)
	mux.HandleApiFunc("/api/v1/parts", api.Parts, models.RoleVVGOVerifiedMember)
	mux.HandleApiFunc("/api/v1/projects", api.Projects, models.RoleAnonymous)
	mux.HandleApiFunc("/api/v1/sessions", api.Sessions, models.RoleVVGOExecutiveDirector)
	mux.HandleFunc("/api/v1/slash_commands", slash_command.Handle, models.RoleAnonymous)
	mux.HandleFunc("/api/v1/slack_commands/list", slash_command.List, models.RoleVVGOProductionTeam)
	mux.HandleFunc("/api/v1/slack_commands/update", slash_command.Update, models.RoleVVGOProductionTeam)
	mux.HandleApiFunc("/api/v1/spreadsheet", api.Spreadsheet, models.RoleWriteSpreadsheet)
	mux.HandleApiFunc("/api/v1/version", api.Version, models.RoleAnonymous)
	mux.HandleApiFunc("/download", api.Download, models.RoleDownload)

	if config.Config.Development {
		mux.HandleFunc("/api/v1/devel/fetch_spreadsheets", devel.FetchSpreadsheets, models.RoleVVGOProductionTeam)
	}

	mux.Handle("/images/", http.FileServer(http.Dir(PublicFiles)), models.RoleAnonymous)
	mux.Handle("/", ServeUI, models.RoleAnonymous)
	return &mux
}
