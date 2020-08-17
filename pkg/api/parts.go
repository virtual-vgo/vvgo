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
	ProjectTitle       string
	PartName           string
	ScoreOrder         int
	SheetMusicFile     string
	ClickTrackFile     string
	ConductorVideo     string
	Released           bool
	Archived           bool
	ReferenceTrack     string
	PronunciationGuide string
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

	parts, err := x.listParts(ctx)
	if err != nil {
		logger.WithError(err).Error("x.Parts.List() failed")
		internalServerError(w)
		return
	}

	parts = x.filterFromQuery(r, parts)
	x.renderView(w, ctx, parts)
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

func (x PartView) listParts(ctx context.Context) ([]Part, error) {
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Sheets client: %w", err)
	}

	readRange := "Parts"
	resp, err := srv.Spreadsheets.Values.Get(x.SpreadSheetID, readRange).Do()
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
		if len(row) < 1 {
			continue
		}
		if len(row) > index["Score Order"] {
			parts[i].ScoreOrder, _ = strconv.Atoi(fmt.Sprint(row[index["Score Order"]]))
		}
		if len(row) > index["Released"] {
			parts[i].Released, _ = strconv.ParseBool(fmt.Sprint(row[index["Released"]]))
		}
		if len(row) > index["Archived"] {
			parts[i].Archived, _ = strconv.ParseBool(fmt.Sprint(row[index["Archived"]]))
		}
		if len(row) > index["Project"] {
			parts[i].Project = fmt.Sprint(row[index["Project"]])
		}
		if len(row) > index["Project Title"] {
			parts[i].ProjectTitle = fmt.Sprint(row[index["Project Title"]])
		}
		if len(row) > index["Part Name"] {
			parts[i].PartName = fmt.Sprint(row[index["Part Name"]])
		}
		if len(row) > index["Sheet Music File"] {
			parts[i].SheetMusicFile = fmt.Sprint(row[index["Sheet Music File"]])
		}
		if len(row) > index["Click Track File"] {
			parts[i].ClickTrackFile = fmt.Sprint(row[index["Click Track File"]])
		}
		if len(row) > index["Conductor Video"] {
			parts[i].ConductorVideo = fmt.Sprint(row[index["Conductor Video"]])
		}
		if len(row) > index["Reference Track"] {

			parts[i].ReferenceTrack = fmt.Sprint(row[index["Reference Track"]])
		}
		if len(row) > index["Pronunciation Guide"] {
			parts[i].PronunciationGuide = fmt.Sprint(row[index["Pronunciation Guide"]])
		}
	}
	return parts, nil
}

func (x PartView) renderView(w http.ResponseWriter, ctx context.Context, parts []Part) {
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
			SheetMusic:         downloadLink(x.Distro.Name, part.SheetMusicFile),
			ClickTrack:         downloadLink(x.Distro.Name, part.ClickTrackFile),
			ReferenceTrack:     downloadLink(x.Distro.Name, part.ReferenceTrack),
			ConductorVideo:     part.ConductorVideo,
			PronunciationGuide: downloadLink(x.Distro.Name, part.PronunciationGuide),
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
