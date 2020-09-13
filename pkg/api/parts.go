package api

import (
	"bytes"
	"context"
	"net/http"
	"path/filepath"
	"strings"
)

type PartView struct {
	SpreadSheetID string
	ReadRange     string
	*Database
}

func (x PartView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	projects, err := listProjects(ctx, x.SpreadSheetID)
	if err != nil {
		logger.WithError(err).Error("x.Parts.List() failed")
		internalServerError(w)
		return
	}

	var wantProjects []Project
	for _, project := range projects {
		if project.Archived == false && project.Released == true {
			wantProjects = append(wantProjects, project)
		}
	}

	parts, err := listParts(ctx, x.SpreadSheetID)
	if err != nil {
		logger.WithError(err).Error("x.Parts.List() failed")
		internalServerError(w)
		return
	}

	renderPartsView(w, ctx, wantProjects, parts, x.Distro.Name)
}

func renderPartsView(w http.ResponseWriter, ctx context.Context, projects []Project, parts []Part, distroBucket string) {
	type tableRow struct {
		Project            string `json:"project"`
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
			Project:            part.ProjectTitle,
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
		NavBar   NavBarOpts
		Rows     []tableRow
		Projects []Project
	}{
		NavBar:   opts,
		Projects: projects,
		Rows:     rows,
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "parts.gohtml")); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}
