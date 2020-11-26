package project

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
)

type Project struct {
	Name                    string
	Title                   string
	Released                bool
	Archived                bool
	Sources                 string
	Composers               string
	Arrangers               string
	Editors                 string
	Transcribers            string
	Preparers               string
	ClixBy                  string `col_name:"Clix By"`
	Reviewers               string
	Lyricists               string
	AdditionalContent       string `col_name:"Additional Content"`
	ReferenceTrack          string `col_name:"Reference Track"`
	ChoirPronunciationGuide string `col_name:"Choir Pronunciation Guide"`
	YoutubeLink             string `col_name:"Youtube Link"`
	YoutubeEmbed            string `col_name:"Youtube Embed"`
	SubmissionDeadline      string `col_name:"Submission Deadline"`
	SubmissionLink          string `col_name:"Submission Link"`
	Season                  string
	BannerLink              string `col_name:"Banner Link"`
}

type Projects []Project

func List(ctx context.Context, identity *login.Identity, spreadsheetID string) (Projects, error) {
	values, err := sheets.ReadSheet(ctx, spreadsheetID, "Projects")
	if err != nil {
		return nil, err
	}
	return valuesToProjects(values).ForIdentity(identity), nil
}

func valuesToProjects(values [][]interface{}) Projects {
	if len(values) < 1 {
		return nil
	}
	index := sheets.BuildIndex(values[0])
	projects := make([]Project, len(values)-1) // ignore the header row
	for i, row := range values[1:] {
		sheets.ProcessRow(row, &projects[i], index)
	}
	return projects
}

func (x Projects) WithName(name string) (Project, bool) {
	for _, project := range x {
		if project.Name == name {
			return project, true
		}
	}
	return Project{}, false
}

func (x Projects) ForIdentity(identity *login.Identity) Projects {
	var want Projects
	for _, project := range x {
		switch {
		case project.Released == true:
			want = append(want, project)
		case identity.HasRole(login.RoleVVGOTeams):
			want = append(want, project)
		case identity.HasRole(login.RoleVVGOLeader):
			want = append(want, project)
		}
	}
	return want
}

func (x Projects) Current() Projects {
	var current []Project
	for _, project := range x {
		if project.Archived == false {
			current = append(current, project)
		}
	}
	return current
}
