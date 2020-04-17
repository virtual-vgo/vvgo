package api

import (
	"bytes"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

type PartsHandler struct {
	NavBar
	*Storage
}

func (x PartsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	type tableRow struct {
		Project      string
		PartName     string
		PartNumber   uint8
		SheetMusic   string
		ClickTrack   string
		BackingTrack string
	}

	allParts := x.Parts.List()
	rows := make([]tableRow, 0, len(allParts))
	for _, part := range allParts {
		rows = append(rows, tableRow{
			Project:      projects.GetName(part.Project).Title,
			PartName:     strings.Title(part.Name),
			PartNumber:   part.Number,
			SheetMusic:   part.SheetLink(x.SheetsBucketName),
			ClickTrack:   part.ClickLink(x.ClixBucketName),
			BackingTrack: projects.GetName(part.Project).BackingTrackLink(x.TracksBucketName),
		})
	}

	navBarOpts := x.NavBar.NewOpts(r)
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
		jsonEncode(&buffer, &rows)
	}
	buffer.WriteTo(w)
}

type IndexHandler struct {
	NavBar
}

func (x IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	navBarOpts := x.NavBar.NewOpts(r)
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

func (x NavBar) NewOpts(r *http.Request) NavBarRenderOpts {
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
