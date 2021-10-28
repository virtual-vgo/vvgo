package models

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
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

type Parts []Part

func ListParts(ctx context.Context, identity Identity) (Parts, error) {
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
		if project, ok := projects.Get(part.Project); ok {
			switch {
			case project.PartsReleased == true && project.PartsArchived == false:
				allowed = append(allowed, part)
			case identity.HasRole(RoleVVGOProductionTeam) && project.PartsArchived == false:
				allowed = append(allowed, part)
			case identity.HasRole(RoleVVGOExecutiveDirector):
				allowed = append(allowed, part)
			}
		}
	}
	return allowed, nil
}

func valuesToParts(values [][]interface{}) Parts {
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

func (x Parts) ForProject(projects ...string) Parts {
	var want Parts
	for _, part := range x {
		for _, project := range projects {
			if part.Project == project {
				want = append(want, part)
			}
		}
	}
	return want
}

func (x Parts) Append(parts Parts) Parts {
	return append(x, parts...)
}

// Sorting

func (x Parts) Len() int           { return len(x) }
func (x Parts) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x Parts) Less(i, j int) bool { return x[i].ScoreOrder < x[j].ScoreOrder }
func (x Parts) Sort() Parts        { sort.Sort(x); return x }
