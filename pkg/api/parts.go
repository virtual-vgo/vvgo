package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type Part struct {
	Project        string
	PartName       string
	ScoreOrder     int
	SheetMusicFile string
	ClickTrackFile string
}

func (x Part) SheetLink(bucket string) string {
	if bucket == "" || x.SheetMusicFile == "" {
		return "#"
	} else {
		return fmt.Sprintf("/download?bucket=%s&object=%s", bucket, x.SheetMusicFile)
	}
}

func (x Part) ClickLink(bucket string) string {
	if bucket == "" || x.ClickTrackFile == "" {
		return "#"
	} else {
		return fmt.Sprintf("/download?bucket=%s&object=%s", bucket, x.ClickTrackFile)
	}
}

type PartView struct {
	SpreadSheetID string
	ReadRange     string
	*Database
}

func (x PartView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "parts_view")
	defer span.Send()

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
		if projects.Exists(parts[i].Project) &&
			projects.GetName(parts[i].Project).Archived == archived &&
			projects.GetName(parts[i].Project).Released == released {
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

	readRange := "Parts!A2:F"
	resp, err := srv.Spreadsheets.Values.Get(x.SpreadSheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data from sheet: %w", err)
	}

	parts := make([]Part, len(resp.Values))
	for i, row := range resp.Values {
		if len(row) != 5 {
			logger.WithField("row", fmt.Sprintf("%#v", row)).Error("invalid columns")
			continue
		}
		scoreOrder, _ := strconv.Atoi(fmt.Sprint(row[2]))
		parts[i] = Part{
			Project:        fmt.Sprint(row[0]),
			PartName:       fmt.Sprint(row[1]),
			ScoreOrder:     scoreOrder,
			SheetMusicFile: fmt.Sprint(row[3]),
			ClickTrackFile: fmt.Sprint(row[4]),
		}
	}
	return parts, nil
}

func (x PartView) renderView(w http.ResponseWriter, ctx context.Context, parts []Part) {
	type tableRow struct {
		Project        string `json:"project"`
		PartName       string `json:"part_name"`
		ScoreOrder     int    `json:"score_order"`
		SheetMusic     string `json:"sheet_music"`
		ClickTrack     string `json:"click_track"`
		ReferenceTrack string `json:"reference_track"`
	}

	rows := make([]tableRow, 0, len(parts))
	for _, part := range parts {
		rows = append(rows, tableRow{
			Project:        projects.GetName(part.Project).Title,
			ScoreOrder:     part.ScoreOrder,
			PartName:       strings.Title(part.PartName),
			SheetMusic:     part.SheetLink(x.Distro.Name),
			ClickTrack:     part.ClickLink(x.Distro.Name),
			ReferenceTrack: projects.GetName(part.Project).ReferenceTrackLink(x.Distro.Name),
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
	buffer.WriteTo(w)
}
