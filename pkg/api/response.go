package api

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api/arrangements"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/credits"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
	auth2 "github.com/virtual-vgo/vvgo/pkg/api/http/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/mixtape"
	"github.com/virtual-vgo/vvgo/pkg/api/traces"
	"github.com/virtual-vgo/vvgo/pkg/api/version"
	"github.com/virtual-vgo/vvgo/pkg/api/website_data"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
)

type Status string

const StatusOk Status = "ok"
const StatusFound Status = "found"
const StatusError Status = "error"

type Response struct {
	Status          Status
	Location        string           `json:"Location,omitempty"`
	Version         *version.Version `json:"Version,omitempty"`
	Error           *errors.Error    `json:"Error,omitempty"`
	Projects        []website_data.Project    `json:"Projects,omitempty"`
	Parts           []website_data.Part       `json:"Parts,omitempty"`
	Sessions        []auth.Identity           `json:"Sessions,omitempty"`
	Spreadsheet     *website_data.Spreadsheet `json:"Spreadsheet,omitempty"`
	Credits         []credits.Credit          `json:"credits,omitempty"`
	Dataset         []map[string]string       `json:"Dataset,omitempty"`
	Identity        *auth.Identity            `json:"Identity,omitempty"`
	GuildMembers    []discord.GuildMember     `json:"GuildMembers,omitempty"`
	Channels        []discord.Channel         `json:"channels,omitempty"`
	MixtapeProjects []mixtape.Project         `json:"MixtapeProjects,omitempty"`
	MixtapeProject  *mixtape.Project          `json:"MixtapeProject,omitempty"`
	CreditsTable    credits.Table             `json:"CreditsTable,omitempty"`
	Ballot          arrangements.Ballot  `json:"Ballot,omitempty"`
	OAuthRedirect   *auth2.OAuthRedirect `json:"OAuthRedirect,omitempty"`
	CreditsPasta    *credits.Pasta       `json:"CreditsPasta,omitempty"`
	Spans           []tracing.Span            `json:"Spans,omitempty"`
	Waterfalls      []traces.Waterfall        `json:"Waterfalls,omitempty"`
}

func NewOkResponse() Response { return Response{Status: StatusOk} }

func (resp Response) WriteHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if resp.Status == StatusFound {
		http.Redirect(w, r, resp.Location, http.StatusFound)
		return
	}

	var code int
	switch {
	case resp.Error != nil:
		code = resp.Error.Code
	case resp.Status == StatusOk:
		code = http.StatusOK
	default:
		code = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.JsonEncodeFailure(ctx, err)
	}
}
