package api

import (
	"bytes"
	"context"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type LoginView struct {
	Sessions *login.Store
}

func (x LoginView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "login_view")
	defer span.Send()

	var identity login.Identity
	if err := x.Sessions.ReadSessionFromRequest(ctx, r, &identity); err == nil && !identity.IsAnonymous() {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	opts := NewNavBarOpts(ctx)
	opts.LoginActive = true
	page := struct {
		NavBar NavBarOpts
	}{
		NavBar: opts,
	}

	var buf bytes.Buffer
	if ok := parseAndExecute(&buf, &page, filepath.Join(PublicFiles, "login.gohtml")); !ok {
		internalServerError(w)
		return
	}
	buf.WriteTo(w)
}

type PartView struct {
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

type IndexView struct{}

func (x IndexView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "index_view")
	defer span.Send()

	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	opts := NewNavBarOpts(ctx)
	page := struct {
		NavBar NavBarOpts
	}{
		NavBar: opts,
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(&buffer, &page, filepath.Join(PublicFiles, "index.gohtml")); !ok {
		internalServerError(w)
		return
	}
	buffer.WriteTo(w)
}

type NavBarOpts struct {
	ShowLogin       bool
	ShowMemberLinks bool
	ShowAdminLinks  bool
	PartsActive     bool
	LoginActive     bool
	BackupsActive   bool
}

func NewNavBarOpts(ctx context.Context) NavBarOpts {
	identity := identityFromContext(ctx)
	return NavBarOpts{
		ShowMemberLinks: identity.HasRole(login.RoleVVGOMember),
		ShowAdminLinks:  identity.HasRole(login.RoleVVGOUploader),
		ShowLogin:       identity.IsAnonymous(),
	}
}

func identityFromContext(ctx context.Context) *login.Identity {
	ctxIdentity := ctx.Value(CtxKeyVVGOIdentity)
	identity, ok := ctxIdentity.(*login.Identity)
	if !ok {
		identity = new(login.Identity)
		*identity = login.Anonymous()
	}
	return identity
}

func parseAndExecute(dest io.Writer, data interface{}, templateFiles ...string) bool {
	templateFiles = append(templateFiles,
		filepath.Join(PublicFiles, "header.gohtml"),
		filepath.Join(PublicFiles, "navbar.gohtml"),
		filepath.Join(PublicFiles, "footer.gohtml"),
	)
	uploadTemplate, err := template.ParseFiles(templateFiles...)
	if err != nil {
		logger.WithError(err).Error("template.ParseFiles() failed")
		return false
	}
	if err := uploadTemplate.Execute(dest, &data); err != nil {
		logger.WithError(err).Error("template.Execute() failed")
		return false
	}
	return true
}
