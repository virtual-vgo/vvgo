package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/redis"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

type Template struct {
	SpreadsheetID string
	DistroBucket  string
}

func getSpreadsheetID(ctx context.Context) string {
	var spreadsheetID string
	redis.Do(ctx, redis.Cmd(&spreadsheetID, "GET", "config:website_data_spreadsheet_id"))
	return spreadsheetID
}

func getDistroBucket(ctx context.Context) string {
	var distroBucket string
	redis.Do(ctx, redis.Cmd(&distroBucket, "GET", "config:distro_bucket"))
	return distroBucket
}

func (x Template) ParseAndExecute(ctx context.Context, w http.ResponseWriter, r *http.Request, data interface{}, templateFile string) {
	x.SpreadsheetID = getSpreadsheetID(ctx)
	x.DistroBucket = getDistroBucket(ctx)

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
		"download_link":    func(obj string) string { return downloadLink(x.DistroBucket, obj) },
		"projects":         func() (sheets.Projects, error) { return sheets.ListProjects(ctx, identity, x.SpreadsheetID) },
		"parts":            func() (sheets.Parts, error) { return sheets.ListParts(ctx, identity, x.SpreadsheetID) },
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

func currentProjects(ctx context.Context, identity *login.Identity, spreadsheetID string) (sheets.Projects, error) {
	projects, err := sheets.ListProjects(ctx, identity, spreadsheetID)
	if err != nil {
		return nil, err
	}
	return projects.Current(), nil
}

func currentParts(ctx context.Context, identity *login.Identity, spreadsheetID string) (sheets.Parts, error) {
	parts, err := sheets.ListParts(ctx, identity, spreadsheetID)
	if err != nil {
		return nil, err
	}
	return parts.Current(), nil
}
