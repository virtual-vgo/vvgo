package api

import (
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api/arrangements"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/channels"
	"github.com/virtual-vgo/vvgo/pkg/api/credits"
	"github.com/virtual-vgo/vvgo/pkg/api/download"
	"github.com/virtual-vgo/vvgo/pkg/api/guild_members"
	"github.com/virtual-vgo/vvgo/pkg/api/me"
	"github.com/virtual-vgo/vvgo/pkg/api/mixtape"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/api/sessions"
	"github.com/virtual-vgo/vvgo/pkg/api/slash_command"
	"github.com/virtual-vgo/vvgo/pkg/api/traces"
	"github.com/virtual-vgo/vvgo/pkg/api/version"
	"github.com/virtual-vgo/vvgo/pkg/api/website_data"
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
	fmt.Println(name)
	if errors.Is(err, os.ErrNotExist) {
		return os.Open(path.Join(PublicFiles, "dist", "index.html"))
	}
	return file, err
}

func authorize(role auth.Role) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		identity := auth.IdentityFromContext(r.Context())
		fmt.Println(identity)
		if !identity.HasRole(role) {
			response.NewUnauthorizedError().WriteHTTP(ctx, w, r)
		}
	}
}

func Routes() http.Handler {
	rbacMux := RBACMux{ServeMux: http.NewServeMux()}

	// authorize
	for _, role := range []auth.Role{auth.RoleVVGOVerifiedMember, auth.RoleVVGOProductionTeam, auth.RoleVVGOExecutiveDirector} {
		rbacMux.HandleFunc("/authorize/"+role.String(), authorize(role), auth.RoleAnonymous)
	}

	// debug
	rbacMux.HandleFunc("/debug/pprof/", pprof.Index, auth.RoleVVGOProductionTeam)
	rbacMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline, auth.RoleVVGOProductionTeam)
	rbacMux.HandleFunc("/debug/pprof/profile", pprof.Profile, auth.RoleVVGOProductionTeam)
	rbacMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol, auth.RoleVVGOProductionTeam)
	rbacMux.HandleFunc("/debug/pprof/trace", pprof.Trace, auth.RoleVVGOProductionTeam)

	// slash commands
	rbacMux.HandleFunc("/api/v1/slash_commands/list", slash_command.List, auth.RoleVVGOProductionTeam)
	rbacMux.HandleFunc("/api/v1/slash_commands/update", slash_command.Update, auth.RoleVVGOProductionTeam)
	rbacMux.HandleFunc("/api/v1/slash_commands", slash_command.Handle, auth.RoleAnonymous)

	// api endpoints
	rbacMux.HandleApiFunc("/api/v1/arrangements/ballot", arrangements.ServeBallot, auth.RoleVVGOExecutiveDirector)
	rbacMux.HandleApiFunc("/api/v1/auth/discord", auth.Discord, auth.RoleAnonymous)
	rbacMux.HandleApiFunc("/api/v1/auth/logout", auth.Logout, auth.RoleAnonymous)
	rbacMux.HandleApiFunc("/api/v1/auth/oauth_redirect", auth.ServeOAuthRedirect, auth.RoleAnonymous)
	rbacMux.HandleApiFunc("/api/v1/auth/password", auth.Password, auth.RoleAnonymous)
	rbacMux.HandleApiFunc("/api/v1/channels/list", channels.HandleList, auth.RoleVVGOVerifiedMember)
	rbacMux.HandleApiFunc("/api/v1/credits", credits.ServeCredits, auth.RoleAnonymous)
	rbacMux.HandleApiFunc("/api/v1/credits/pasta", credits.ServePasta, auth.RoleVVGOProductionTeam)
	rbacMux.HandleApiFunc("/api/v1/credits/table", credits.ServeTable, auth.RoleAnonymous)
	rbacMux.HandleApiFunc("/api/v1/dataset", website_data.ServeDataset, auth.RoleAnonymous)
	rbacMux.HandleApiFunc("/api/v1/download", download.Download, auth.RoleDownload)
	rbacMux.HandleApiFunc("/api/v1/guild_members/list", guild_members.HandleList, auth.RoleVVGOVerifiedMember)
	rbacMux.HandleApiFunc("/api/v1/guild_members/lookup", guild_members.HandleLookup, auth.RoleVVGOVerifiedMember)
	rbacMux.HandleApiFunc("/api/v1/guild_members/search", guild_members.HandleSearch, auth.RoleVVGOVerifiedMember)
	rbacMux.HandleApiFunc("/api/v1/me", me.Me, auth.RoleAnonymous)
	rbacMux.HandleApiFunc("/api/v1/mixtape/projects/", mixtape.ServeProjects, auth.RoleVVGOVerifiedMember)
	rbacMux.HandleApiFunc("/api/v1/parts", website_data.ServeParts, auth.RoleVVGOVerifiedMember)
	rbacMux.HandleApiFunc("/api/v1/projects", website_data.ServeProjects, auth.RoleAnonymous)
	rbacMux.HandleApiFunc("/api/v1/sessions", sessions.Sessions, auth.RoleVVGOVerifiedMember)
	rbacMux.HandleApiFunc("/api/v1/spreadsheet", website_data.ServeSpreadsheet, auth.RoleWriteSpreadsheet)
	rbacMux.HandleApiFunc("/api/v1/traces/spans", traces.ServeSpans, auth.RoleVVGOProductionTeam)
	rbacMux.HandleApiFunc("/api/v1/traces/waterfall", traces.ServeWaterfall, auth.RoleVVGOExecutiveDirector)
	rbacMux.HandleApiFunc("/api/v1/version", version.Version, auth.RoleAnonymous)
	rbacMux.HandleApiFunc("/download", download.Download, auth.RoleDownload)

	rbacMux.Handle("/images/", http.FileServer(http.Dir(PublicFiles)), auth.RoleAnonymous)
	rbacMux.Handle("/", ServeUI, auth.RoleAnonymous)
	return &rbacMux
}
