package website_data

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/auth"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"sort"
)

const SheetParts = "Parts"

type Part struct {
	Project            string
	PartName           string
	ScoreOrder         int
	SheetMusicFile     string
	ClickTrackFile     string
	ConductorVideo     string
	PronunciationGuide string

	// Derived Columns
	SheetMusicLink         string
	ClickTrackLink         string
	PronunciationGuideLink string
}

func ServeParts(r *http.Request) api.Response {
	ctx := r.Context()
	identity := auth.IdentityFromContext(ctx)

	parts, err := ListParts(ctx, identity)
	if err != nil {
		logger.ListPartsFailure(ctx, err)
		return response.NewInternalServerError()
	}

	if parts == nil {
		parts = []Part{}
	}
	sort.Slice(parts, func(i, j int) bool { return parts[i].ScoreOrder < parts[j].ScoreOrder })
	return api.Response{Status: api.StatusOk, Parts: parts}
}

func ListParts(ctx context.Context, identity auth.Identity) ([]Part, error) {
	if identity.IsAnonymous() {
		return []Part{}, nil
	}

	projects, err := ListProjects(ctx, identity)
	if err != nil {
		return nil, err
	}

	values, err := redis.ReadSheet(ctx, SpreadsheetWebsiteData, SheetParts)
	if err != nil {
		return nil, err
	}
	parts := valuesToParts(values)

	var allowed []Part
	for _, part := range parts {
		if project, ok := GetProject(projects, part.Project); ok {
			switch {
			case project.IsAcceptingSubmissions():
				allowed = append(allowed, part)
			case identity.HasRole(auth.RoleVVGOProductionTeam) && project.PartsArchived == false:
				allowed = append(allowed, part)
			case identity.HasRole(auth.RoleVVGOExecutiveDirector):
				allowed = append(allowed, part)
			}
		}
	}
	return allowed, nil
}

func valuesToParts(values [][]interface{}) []Part {
	if len(values) < 1 {
		return nil
	}
	parts := make([]Part, 0, len(values)-1)
	UnmarshalSheet(values, &parts)
	for i := range parts {
		if parts[i].SheetMusicLink == "" {
			parts[i].SheetMusicLink = downloadLink(parts[i].SheetMusicFile)
		}
		if parts[i].ClickTrackLink == "" {
			parts[i].ClickTrackLink = downloadLink(parts[i].ClickTrackFile)
		}
		if parts[i].PronunciationGuideLink == "" {
			parts[i].PronunciationGuideLink = downloadLink(parts[i].PronunciationGuide)
		}
	}
	return parts
}

func downloadLink(object string) string {
	if object == "" {
		return ""
	} else {
		return "/download?object=" + object
	}
}
