package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

func ParseAndExecute(ctx context.Context, w http.ResponseWriter, r *http.Request, data interface{}, templateFile string) {
	identity := IdentityFromContext(ctx)

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(map[string]interface{}{
		"template_file":    func() string { return templateFile },
		"link_to_template": func() string { return "https://github.com/virtual-vgo/vvgo/blob/master/public/" + templateFile },
		"user_info":        identity.Info,
		"user_roles":       func() []login.Role { return identity.Roles },
		"title":            strings.Title,
		"form_value":       func(key string) string { return r.FormValue(key) },
		"user_logged_in":   func() bool { return identity.IsAnonymous() == false },
		"user_is_member":   func() bool { return identity.HasRole(login.RoleVVGOMember) },
		"user_is_leader":   func() bool { return identity.HasRole(login.RoleVVGOLeader) },
		"user_on_teams":    func() bool { return identity.HasRole(login.RoleVVGOTeams) },
		"download_link":    func(obj string) string { return downloadLink(obj) },
		"projects":         func() (sheets.Projects, error) { return sheets.ListProjects(ctx, identity) },
		"parts":            func() (sheets.Parts, error) { return sheets.ListParts(ctx, identity) },
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

func downloadLink(object string) string {
	if object == "" {
		return "#"
	} else {
		return fmt.Sprintf("/download?object=%s", object)
	}
}
