package api

import (
	"bytes"
	"context"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type ArchiveView struct {
	SpreadSheetID string
	*Database
}

func (x ArchiveView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/archive/" {
		x.serveIndex(w, r)
	} else {
		x.serveProject(w, r, r.URL.Path[len("/archive/"):])
	}
}

func (x ArchiveView) serveIndex(w http.ResponseWriter, r *http.Request) {
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

func (x ArchiveView) filterFromQuery(r *http.Request, projects []Project) []Project {
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

func (x ArchiveView) renderIndexView(w http.ResponseWriter, ctx context.Context, projects []Project) {
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
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "archive/index.gohtml")); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}

func (x ArchiveView) serveProject(w http.ResponseWriter, r *http.Request, name string) {
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
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "archive/project.gohtml")); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}

