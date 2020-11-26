package api

import (
	"bytes"
	"context"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/models/part"
	"github.com/virtual-vgo/vvgo/pkg/models/project"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

type IndexView struct{}

func (x IndexView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(ctx, &buffer, &struct{}{}, "index.gohtml"); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}

type AboutView struct {
	SpreadSheetID string
}

func (x AboutView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	leaders, err := models.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("readSheet() failed")
		internalServerError(w)
		return
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(ctx, &buffer, leaders, "about.gohtml"); !ok {
		internalServerError(w)
		return
	}
	_, _ = buffer.WriteTo(w)
}

func parseAndExecute(ctx context.Context, dest io.Writer, data interface{}, templateFile string) bool {
	identity := IdentityFromContext(ctx)

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(map[string]interface{}{
		"link_to_template": func() string { return "https://github.com/virtual-vgo/vvgo/blob/master/public/" + templateFile },
		"user_info":        identity.Info,
		"user_logged_in":   func() bool { return identity.IsAnonymous() },
		"user_is_member":   func() bool { return identity.HasRole(login.RoleVVGOMember) },
		"user_is_leader":   func() bool { return identity.HasRole(login.RoleVVGOLeader) },
		"user_on_teams":    func() bool { return identity.HasRole(login.RoleVVGOTeams) },
		"template_file":    func() string { return templateFile },
		"projects":         func() ([]project.Project, error) { return models.ListProjects(ctx, identity) },
		"current_projects": func() ([]project.Project, error) { return models.ListCurrentProjects(ctx, identity) },
		"parts":            func() ([]part.Part, error) { return models.ListParts(ctx, identity) },
		"current_parts":    func() ([]part.Part, error) { return models.ListCurrentParts(ctx, identity) },
		"download_link":    func(obj string) string { return downloadLink("vvgo-distro", obj) },
		"title":            strings.Title,
	}).ParseFiles(
		PublicFiles+"/"+templateFile,
		PublicFiles+"/header.gohtml",
		PublicFiles+"/navbar.gohtml",
		PublicFiles+"/footer.gohtml",
	)
	if err != nil {
		logger.WithError(err).Error("template.ParseFiles() failed")
		return false
	}

	if err := tmpl.Execute(dest, &data); err != nil {
		logger.WithError(err).Error("template.Execute() failed")
		return false
	}
	return true
}
