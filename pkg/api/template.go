package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/sheets/part"
	"github.com/virtual-vgo/vvgo/pkg/sheets/project"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

type Template struct {
	SpreadsheetID string
	DistroBucket  string
}

func (x Template) ParseAndExecute(ctx context.Context, w http.ResponseWriter, r *http.Request, data interface{}, templateFile string) {
	identity := IdentityFromContext(ctx)

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(map[string]interface{}{
		"template_file":    func() string { return templateFile },
		"link_to_template": func() string { return "https://github.com/virtual-vgo/vvgo/blob/master/public/" + templateFile },
		"user_info":        identity.Info,
		"title":            strings.Title,
		"form_value":       func(key string) string { return r.FormValue(key) },
		"user_logged_in":   func() bool { return identity.IsAnonymous() == false },
		"user_is_member":   func() bool { return identity.HasRole(login.RoleVVGOMember) },
		"user_is_leader":   func() bool { return identity.HasRole(login.RoleVVGOLeader) },
		"user_on_teams":    func() bool { return identity.HasRole(login.RoleVVGOTeams) },
		"download_link":    func(obj string) string { return downloadLink(x.DistroBucket, obj) },
		"projects":         func() ([]project.Project, error) { return project.List(ctx, identity, x.SpreadsheetID) },
		"current_projects": func() ([]project.Project, error) { return currentProjects(ctx, identity, x.SpreadsheetID) },
		"parts":            func() ([]part.Part, error) { return part.List(ctx, identity, x.SpreadsheetID) },
		"current_parts":    func() ([]part.Part, error) { return currentParts(ctx, identity, x.SpreadsheetID) },
	}).ParseFiles(
		PublicFiles+"/"+templateFile,
		PublicFiles+"/header.gohtml",
		PublicFiles+"/navbar.gohtml",
		PublicFiles+"/footer.gohtml",
	)

	if err != nil {
		logger.WithError(err).Error("template.ParseFiles() failed")
		internalServerError(w)
		return
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, &data); err != nil {
		logger.WithError(err).Error("template.Execute() failed")
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

func currentProjects(ctx context.Context, identity *login.Identity, spreadsheetID string) (project.Projects, error) {
	projects, err := project.List(ctx, identity, spreadsheetID)
	if err != nil {
		return nil, err
	}
	return projects.Current(), nil
}

func currentParts(ctx context.Context, identity *login.Identity, spreadsheetID string) (part.Parts, error) {
	parts, err := part.List(ctx, identity, spreadsheetID)
	if err != nil {
		return nil, err
	}
	return parts.Current(), nil
}
