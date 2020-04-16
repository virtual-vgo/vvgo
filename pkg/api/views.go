package api

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

type PartsHandler struct{ *Storage }

func (x PartsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	type tableRow struct {
		Project    string `json:"project"`
		PartName   string `json:"part_name"`
		PartNumber uint8  `json:"part_number"`
		SheetMusic string `json:"sheet_music"`
		ClickTrack string `json:"click_track"`
	}

	allParts := x.Parts.List()
	rows := make([]tableRow, 0, len(allParts))
	for _, part := range allParts {
		rows = append(rows, tableRow{
			Project:    part.Project,
			PartName:   strings.Title(part.Name),
			PartNumber: part.Number,
			SheetMusic: part.SheetLink(x.SheetsBucketName),
			ClickTrack: part.ClickLink(x.ClixBucketName),
		})
	}

	page := struct {
		Header template.HTML
		NavBar template.HTML
		Rows   []tableRow
	}{
		Header: Header(),
		NavBar: NavBar(NavBarOpts{PartsActive: true}),
		Rows:   rows,
	}

	var buffer bytes.Buffer
	switch true {
	case acceptsType(r, "text/html"):
		if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "parts.gohtml")); !ok {
			internalServerError(w)
			return
		}
	default:
		jsonEncode(&buffer, &rows)
	}
	buffer.WriteTo(w)
}

type IndexHandler struct{}

func (x IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	page := struct {
		Header template.HTML
		NavBar template.HTML
	}{
		Header: Header(),
		NavBar: NavBar(NavBarOpts{}),
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "index.gohtml")); !ok {
		internalServerError(w)
		return
	}
	buffer.WriteTo(w)
}

func Header() template.HTML {
	var buffer bytes.Buffer
	parseAndExecute(&buffer, &struct{}{}, filepath.Join(PublicFiles, "header.gohtml"))
	return template.HTML(buffer.String())
}

func NavBar(opts NavBarOpts) template.HTML {
	var buffer bytes.Buffer
	parseAndExecute(&buffer, &opts, filepath.Join(PublicFiles, "navbar.gohtml"))
	return template.HTML(buffer.String())
}

type NavBarOpts struct {
	PartsActive bool
}
