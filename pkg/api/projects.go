package api

import (
	"bytes"
	"context"
	"fmt"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"path/filepath"
	"reflect"
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
		processRow(row, &projects[i], index)
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
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "projects/index.gohtml")); !ok {
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
	page := struct {
		NavBar NavBarOpts
		Project
		Rows []tableRow
	}{
		NavBar:  opts,
		Project: project,
		Rows:    rows,
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "projects/project.gohtml")); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}

func processRow(row []interface{}, dest interface{}, index map[string]int) {
	tagName := "col_name"
	if len(row) < 1 {
		return
	}
	reflectType := reflect.TypeOf(dest).Elem()
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		colName := field.Tag.Get(tagName)
		if colName == "" {
			colName = field.Name
		}
		colIndex, ok := index[colName]
		if !ok {
			continue
		}
		if len(row) > colIndex {
			switch field.Type.Kind() {
			case reflect.String:
				val := fmt.Sprint(row[colIndex])
				reflect.ValueOf(dest).Elem().Field(i).SetString(val)
			case reflect.Bool:
				val, _ := strconv.ParseBool(fmt.Sprint(row[colIndex]))
				reflect.ValueOf(dest).Elem().Field(i).SetBool(val)
			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				val, _ := strconv.ParseInt(fmt.Sprint(row[colIndex]), 10, 64)
				reflect.ValueOf(dest).Elem().Field(i).SetInt(val)
			}
		}
	}
}
