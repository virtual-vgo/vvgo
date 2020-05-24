package api

import (
	"bytes"
	"context"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type PartView struct {
	NavBar
	*Database
}

func (x PartView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "parts_view")
	defer span.Send()

	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	type tableRow struct {
		Project        string `json:"project"`
		PartName       string `json:"part_name"`
		PartNumber     uint8  `json:"part_number"`
		SheetMusic     string `json:"sheet_music"`
		ClickTrack     string `json:"click_track"`
		ReferenceTrack string `json:"reference_track"`
	}

	parts, err := x.Parts.List(ctx)
	if err != nil {
		logger.WithError(err).Error("x.Parts.List() failed")
		internalServerError(w)
		return
	}

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
		if parts[i].Validate() == nil &&
			projects.GetName(parts[i].Project).Archived == archived &&
			projects.GetName(parts[i].Project).Released == released {
			continue
		}
		parts[i], parts[want-1] = parts[want-1], parts[i]
		i--
		want--
	}
	parts = parts[:want]
	rows := make([]tableRow, 0, len(parts))
	for _, part := range parts {
		rows = append(rows, tableRow{
			Project:        projects.GetName(part.Project).Title,
			PartName:       strings.Title(part.Name),
			PartNumber:     part.Number,
			SheetMusic:     part.SheetLink(x.Distro.Name),
			ClickTrack:     part.ClickLink(x.Distro.Name),
			ReferenceTrack: projects.GetName(part.Project).ReferenceTrackLink(x.Distro.Name),
		})
	}

	navBarOpts := x.NavBar.NewOpts(ctx, r)
	navBarOpts.PartsActive = true
	page := struct {
		Header template.HTML
		NavBar template.HTML
		Rows   []tableRow
	}{
		Header: Header(),
		NavBar: x.NavBar.RenderHTML(navBarOpts),
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
		jsonEncodeBeautify(&buffer, &rows)
	}
	buffer.WriteTo(w)
}

type IndexView struct {
	NavBar
}

func (x IndexView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "index_view")
	defer span.Send()

	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	navBarOpts := x.NavBar.NewOpts(ctx, r)
	page := struct {
		Header template.HTML
		NavBar template.HTML
	}{
		Header: Header(),
		NavBar: x.NavBar.RenderHTML(navBarOpts),
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

type NavBar struct {
	MemberUser string
}

type NavBarRenderOpts struct {
	ShowLogin       bool
	ShowMemberLinks bool
	PartsActive     bool
}

func (x NavBar) NewOpts(ctx context.Context, r *http.Request) NavBarRenderOpts {
	var opts NavBarRenderOpts
	user, _, _ := r.BasicAuth()
	switch user {
	case x.MemberUser:
		opts.ShowMemberLinks = true
	default:
		opts.ShowLogin = true
	}
	return opts
}

func (x NavBar) RenderHTML(opts NavBarRenderOpts) template.HTML {
	var buffer bytes.Buffer
	parseAndExecute(&buffer, &opts, filepath.Join(PublicFiles, "navbar.gohtml"))
	return template.HTML(buffer.String())
}
