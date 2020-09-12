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

	parts, err := listParts(ctx, x.SpreadSheetID)
	if err != nil {
		logger.WithError(err).Error("x.Parts.List() failed")
		internalServerError(w)
		return
	}

	parts = x.filterFromQuery(r, parts)
	renderPartsView(w, ctx, parts, x.Distro.Name)
}

func (x PartView) filterFromQuery(r *http.Request, parts []Part) []Part {
	archived := false
	released := true

	if want := r.FormValue("archived"); want != "" {
		archived, _ = strconv.ParseBool(want)
	}

	if want := r.FormValue("released"); want != "" {
		released, _ = strconv.ParseBool(want)
	}

	want := len(parts)
	for i := 0; i < want; i++ {
		if parts[i].Archived == archived &&
			parts[i].Released == released {
			continue
		}
		parts[i], parts[want-1] = parts[want-1], parts[i]
		i--
		want--
	}
	parts = parts[:want]
	return parts
}

func listParts(ctx context.Context, spreadSheetID string) ([]Part, error) {
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Sheets client: %w", err)
	}

	readRange := "Parts"
	resp, err := srv.Spreadsheets.Values.Get(spreadSheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data from sheet: %w", err)
	}

	if len(resp.Values) < 1 {
		return nil, fmt.Errorf("no data")
	}
	parts := make([]Part, len(resp.Values)-1)

	index := make(map[string]int, len(resp.Values[0])-1)
	for i, col := range resp.Values[0] {
		index[fmt.Sprintf("%s", col)] = i
	}

	for i, row := range resp.Values[1:] {
		processRow(row, &parts[i], index)
	}
	return parts, nil
}

func renderPartsView(w http.ResponseWriter, ctx context.Context, parts []Part, distroBucket string) {
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
			Project:            strings.Title(part.ProjectTitle),
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
		Rows   []tableRow
	}{
		NavBar: opts,
		Rows:   rows,
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "parts.gohtml")); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}

func downloadLink(bucket, object string) string {
	if bucket == "" || object == "" {
		return ""
	} else {
		return fmt.Sprintf("/download?bucket=%s&object=%s", bucket, object)
	}
}
