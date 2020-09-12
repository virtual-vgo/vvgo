package api

import (
	"bytes"
	"context"
	"fmt"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"path/filepath"
	"strconv"
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
	ClixBy                  string
	Reviewers               string
	Lyricists               string
	AdditionalContent       string
	ReferenceTrack          string
	ChoirPronunciationGuide string
	YoutubeLink             string
	SubmissionDeadline      string
	SubmissionLink          string
	Season                  string
	BannerLink              string
}

type ProjectsView struct {
	SpreadSheetID string
	*Database
}

func (x ProjectsView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	projects, err := x.listProjects(ctx)
	if err != nil {
		logger.WithError(err).Error("x.listProjects() failed")
		internalServerError(w)
		return
	}

	projects = x.filterFromQuery(r, projects)
	x.renderView(w, ctx, projects)
}

func (x ProjectsView) listProjects(ctx context.Context) ([]Project, error) {
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Sheets client: %w", err)
	}

	readRange := "Projects"
	resp, err := srv.Spreadsheets.Values.Get(x.SpreadSheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data from sheet: %w", err)
	}

	if len(resp.Values) < 1 {
		return nil, fmt.Errorf("no data")
	}
	projects := make([]Project, len(resp.Values)-1) // ignore the header row

	index := make(map[string]int, len(resp.Values[0])-1)
	for i, col := range resp.Values[0] {
		index[fmt.Sprintf("%s", col)] = i
	}

	for i, row := range resp.Values[1:] {
		if len(row) < 1 {
			continue
		}
		if len(row) > index["Name"] {
			projects[i].Name = fmt.Sprint(row[index["Name"]])
		}
		if len(row) > index["Title"] {
			projects[i].Title = fmt.Sprint(row[index["Title"]])
		}
		if len(row) > index["Released"] {
			projects[i].Released, _ = strconv.ParseBool(fmt.Sprint(row[index["Released"]]))
		}
		if len(row) > index["Archived"] {
			projects[i].Archived, _ = strconv.ParseBool(fmt.Sprint(row[index["Archived"]]))
		}
		if len(row) > index["Sources"] {
			projects[i].Sources = fmt.Sprint(row[index["Sources"]])
		}
		if len(row) > index["Composers"] {
			projects[i].Composers = fmt.Sprint(row[index["Composers"]])
		}
		if len(row) > index["Arrangers"] {
			projects[i].Arrangers = fmt.Sprint(row[index["Arrangers"]])
		}
		if len(row) > index["Editors"] {
			projects[i].Editors = fmt.Sprint(row[index["Editors"]])
		}
		if len(row) > index["Transcribers"] {
			projects[i].Transcribers = fmt.Sprint(row[index["Transcribers"]])
		}
		if len(row) > index["Preparers"] {
			projects[i].Preparers = fmt.Sprint(row[index["Preparers"]])
		}
		if len(row) > index["Clix By"] {
			projects[i].ClixBy = fmt.Sprint(row[index["Clix By"]])
		}
		if len(row) > index["Reviewers"] {
			projects[i].Reviewers = fmt.Sprint(row[index["Reviewers"]])
		}
		if len(row) > index["Lyricists"] {
			projects[i].Lyricists = fmt.Sprint(row[index["Lyricists"]])
		}
		if len(row) > index["Additional Content"] {
			projects[i].AdditionalContent = fmt.Sprint(row[index["Additional Content"]])
		}
		if len(row) > index["Reference Track"] {
			projects[i].ReferenceTrack = fmt.Sprint(row[index["Reference Track"]])
		}
		if len(row) > index["Choir Pronunciation Guide"] {
			projects[i].ChoirPronunciationGuide = fmt.Sprint(row[index["Choir Pronunciation Guide"]])
		}
		if len(row) > index["Youtube Link"] {
			projects[i].YoutubeLink = fmt.Sprint(row[index["Youtube Link"]])
		}
		if len(row) > index["Submission Link"] {
			projects[i].SubmissionLink = fmt.Sprint(row[index["Submission Link"]])
		}
		if len(row) > index["Submission Deadline"] {
			projects[i].SubmissionDeadline = fmt.Sprint(row[index["Submission Deadline"]])
		}
		if len(row) > index["Season"] {
			projects[i].Season = fmt.Sprint(row[index["Season"]])
		}
		if len(row) > index["Banner Link"] {
			projects[i].BannerLink = fmt.Sprint(row[index["Banner Link"]])
		}
	}
	return projects, nil
}

func (x ProjectsView) filterFromQuery(r *http.Request, projects []Project) []Project {
	showAll := false

	if want := r.FormValue("showAll"); want != "" {
		showAll, _ = strconv.ParseBool(want)
	}

	want := len(projects)
	for i := 0; i < want; i++ {
		if projects[i].Released == true || showAll {
			continue
		}
		projects[i], projects[want-1] = projects[want-1], projects[i]
		i--
		want--
	}
	projects = projects[:want]
	return projects
}

func (x ProjectsView) renderView(w http.ResponseWriter, ctx context.Context, projects []Project) {
	opts := NewNavBarOpts(ctx)
	opts.ProjectsActive = true
	page := struct {
		NavBar NavBarOpts
		Rows   []Project
	}{
		NavBar: opts,
		Rows:   projects,
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "projects.gohtml")); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}
