package models

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/discord"
	"github.com/virtual-vgo/vvgo/pkg/models/mixtape"
	"github.com/virtual-vgo/vvgo/pkg/models/traces"
	"github.com/virtual-vgo/vvgo/pkg/version"
)

type ApiResponseStatus string

const StatusOk ApiResponseStatus = "ok"
const StatusFound ApiResponseStatus = "found"
const StatusError ApiResponseStatus = "error"

type ApiResponse struct {
	Status          ApiResponseStatus
	Location        string                `json:"Location,omitempty"`
	Version         *version.Version      `json:"Version,omitempty"`
	Error           *ApiError             `json:"Error,omitempty"`
	Projects        []Project             `json:"Projects,omitempty"`
	Parts           []Part                `json:"Parts,omitempty"`
	Sessions        []Identity            `json:"Sessions,omitempty"`
	Spreadsheet     *Spreadsheet          `json:"Spreadsheet,omitempty"`
	Credits         []Credit              `json:"credits,omitempty"`
	Dataset         []map[string]string   `json:"Dataset,omitempty"`
	Identity        *Identity             `json:"Identity,omitempty"`
	GuildMembers    []discord.GuildMember `json:"GuildMembers,omitempty"`
	Channels        []discord.Channel     `json:"channels,omitempty"`
	MixtapeProjects []mixtape.Project     `json:"MixtapeProjects,omitempty"`
	MixtapeProject  *mixtape.Project      `json:"MixtapeProject,omitempty"`
	WorkflowResult  []WorkflowTaskResult  `json:"WorkflowResult,omitempty"`
	CreditsTable    CreditsTable          `json:"CreditsTable,omitempty"`
	Ballot          ArrangementsBallot    `json:"Ballot,omitempty"`
	OAuthRedirect   *OAuthRedirect        `json:"OAuthRedirect,omitempty"`
	CreditsPasta    *CreditsPasta         `json:"CreditsPasta,omitempty"`
	Spans           []traces.Span         `json:"Spans,omitempty"`
	Waterfalls      []traces.Waterfall    `json:"Waterfalls,omitempty"`
}

type ApiError struct {
	Code  int             `json:"Code"`
	Error string          `json:"Error"`
	Data  json.RawMessage `json:"Data"`
}

type CreditsPasta struct {
	WebsitePasta string
	VideoPasta   string
	YoutubePasta string
}

type ArrangementsBallot []string

type OAuthRedirect struct {
	DiscordURL string
	State      string
	Secret     string
}
