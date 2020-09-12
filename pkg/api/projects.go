package api

import (
	"bytes"
	"context"
	"fmt"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
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
	if r.URL.Path == "/projects/" {
		x.serveIndex(w, r)
	} else {
		x.serveProject(w, r, r.URL.Path[len("/projects/"):])
	}
}

func (x ProjectsView) serveIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	projects, err := listProjects(ctx, x.SpreadSheetID)
	if err != nil {
		logger.WithError(err).Error("x.listProjects() failed")
		internalServerError(w)
		return
	}

	projects = x.filterFromQuery(r, projects)
	x.renderIndexView(w, ctx, projects)
}

func listProjects(ctx context.Context, spreadSheetID string) ([]Project, error) {
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Sheets client: %w", err)
	}

	readRange := "Projects"
	resp, err := srv.Spreadsheets.Values.Get(spreadSheetID, readRange).Do()
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

func (x ProjectsView) renderIndexView(w http.ResponseWriter, ctx context.Context, projects []Project) {
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
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "project_index.gohtml")); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}

func (x ProjectsView) serveProject(w http.ResponseWriter, r *http.Request, name string) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	projects, err := listProjects(ctx, x.SpreadSheetID)
	if err != nil {
		logger.WithError(err).Error("x.listProjects() failed")
		internalServerError(w)
		return
	}

	var exists bool
	var wantProject Project
	for _, project := range projects {
		if project.Name == name {
			exists = true
			wantProject = project
			break
		}
	}
	if !exists {
		http.NotFound(w, r)
		return
	}

	parts, err := listParts(ctx, x.SpreadSheetID)
	if err != nil {
		logger.WithError(err).Error("x.Parts.List() failed")
		internalServerError(w)
		return
	}

	var wantParts []Part
	for _, part := range parts {
		if part.Project == name {
			wantParts = append(wantParts, part)
		}
	}

	renderProjectView(w, ctx, wantProject, wantParts, x.Distro.Name)
}

func renderProjectView(w http.ResponseWriter, ctx context.Context, project Project, parts []Part, distroBucket string) {
	type tableRow struct {
		PartName           string `json:"part_name"`
		ScoreOrder         int    `json:"score_order"`
		SheetMusic         string `json:"sheet_music,omitempty"`
		ClickTrack         string `json:"click_track,omitempty"`
		ReferenceTrack     string `json:"reference_track,omitempty"`
		ConductorVideo     string `json:"conductor_video,omitempty"`
		PronunciationGuide string `json:"pronunciation_guide,omitempty"`
	}

	rows := make([]tableRow, 0, len(parts))
	for _, part := range parts {
		rows = append(rows, tableRow{
			ScoreOrder:         part.ScoreOrder,
			PartName:           strings.Title(part.PartName),
			SheetMusic:         downloadLink(distroBucket, part.SheetMusicFile),
			ClickTrack:         downloadLink(distroBucket, part.ClickTrackFile),
			ReferenceTrack:     downloadLink(distroBucket, part.ReferenceTrack),
			ConductorVideo:     part.ConductorVideo,
			PronunciationGuide: downloadLink(distroBucket, part.PronunciationGuide),
		})
	}

	opts := NewNavBarOpts(ctx)
	opts.PartsActive = true
	page := struct {
		NavBar NavBarOpts
		Project
		Rows   []tableRow
	}{
		NavBar: opts,
		Project: project,
		Rows:   rows,
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "project.gohtml")); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}
