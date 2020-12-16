package sheets

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/login"
)

type Part struct {
	Project            string
	ProjectTitle       string `col_name:"Project Title"`
	PartName           string `col_name:"Part Name"`
	ScoreOrder         int    `col_name:"Score Order"`
	SheetMusicFile     string `col_name:"Sheet Music File"`
	ClickTrackFile     string `col_name:"Click Track File"`
	ConductorVideo     string `col_name:"Conductor Video"`
	Released           bool
	Archived           bool
	ReferenceTrack     string `col_name:"Reference Track"`
	PronunciationGuide string `col_name:"Pronunciation Guide"`
}

type Parts []Part

func ListParts(ctx context.Context, identity *login.Identity) (Parts, error) {
	values, err := ReadSheet(ctx, WebsiteDataSpreadsheetID(ctx), "Parts")
	if err != nil {
		return nil, err
	}
	return valuesToParts(values).ForIdentity(identity), nil
}

func valuesToParts(values [][]interface{}) Parts {
	if len(values) < 1 {
		return nil
	}
	index := buildIndex(values[0])
	parts := make([]Part, len(values)-1)
	for i, row := range values[1:] {
		processRow(row, &parts[i], index)
	}
	return parts
}

func (x Parts) ForIdentity(identity *login.Identity) Parts {
	var want Parts
	for _, part := range x {
		switch {
		case part.Released == true && identity.HasRole(login.RoleVVGOMember):
			want = append(want, part)
		case identity.HasRole(login.RoleVVGOTeams):
			want = append(want, part)
		case identity.HasRole(login.RoleVVGOLeader):
			want = append(want, part)
		}
	}
	return want
}

func (x Parts) ForProject(project string) Parts {
	var want Parts
	for _, part := range x {
		if part.Project == project {
			want = append(want, part)
		}
	}
	return want
}

func (x Parts) Current() Parts {
	var current []Part
	for _, part := range x {
		if part.Archived == false {
			current = append(current, part)
		}
	}
	return current
}

func (x Parts) Append(parts Parts) Parts {
	return append(x, parts...)
}
