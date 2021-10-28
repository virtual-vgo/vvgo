package views

import (
	"bytes"
	"context"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var PublicFiles = "public"

var (
	ReactUI            = ServeUI
	ServePublicFile    = http.FileServer(http.Dir(PublicFiles)).ServeHTTP
)

type Page struct {
	Title    string
	JsSource string
	Content  template.HTML
}

func ServeUI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	file, err := os.Open("public/index.html")
	if err != nil {
		logger.OpenFileFailure(ctx, err)
		http_helpers.WriteInternalServerError(ctx, w)
		return
	}
	if _, err := io.Copy(w, file); err != nil {
		logger.MethodFailure(ctx, "io.Copy", err)
	}
}

func (x Page) Render(w http.ResponseWriter, r *http.Request) {
	ParseAndExecute(r.Context(), w, r, x, "page.gohtml")
}

func ParseAndExecute(ctx context.Context, w http.ResponseWriter, r *http.Request, data interface{}, templateFile string) {
	identity := login.IdentityFromContext(ctx)

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(map[string]interface{}{
		"template_file":    func() string { return templateFile },
		"link_to_template": func() string { return "https://github.com/virtual-vgo/vvgo/blob/master/public/" + templateFile },
		"user_info":        identity.Info,
		"user_roles":       func() []models.Role { return identity.Roles },
		"user_identity":    func() models.Identity { return identity },
		"title":            strings.Title,
		"form_value":       func(key string) string { return r.FormValue(key) },
		"user_logged_in":   func() bool { return identity.IsAnonymous() == false },
		"user_is_member":   func() bool { return identity.HasRole(models.RoleVVGOVerifiedMember) },
		"user_is_leader":   func() bool { return identity.HasRole(models.RoleVVGOExecutiveDirector) },
		"user_on_teams":    func() bool { return identity.HasRole(models.RoleVVGOProductionTeam) },
		"download_link":    func(obj string) string { return downloadLink(obj) },
		"projects":         func() (models.Projects, error) { return models.ListProjects(ctx, identity) },
		"parts":            func() (models.Parts, error) { return models.ListParts(ctx, identity) },
		"new_query":        models.NewQuery,
		"string_slice":     func(strs ...string) []string { return strs },
		"append_strings":   func(slice []string, elems ...string) []string { return append(slice, elems...) },
		"pick_random_elem": func(slice []string) string { return slice[rand.Intn(len(slice))] },
	}).ParseFiles(
		PublicFiles+"/"+templateFile,
		PublicFiles+"/header.gohtml",
		PublicFiles+"/navbar.gohtml",
		PublicFiles+"/footer.gohtml",
	)

	if err != nil {
		logger.MethodFailure(ctx, "template.ParseFiles", err)
		http_helpers.WriteInternalServerError(ctx, w)
		return
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, &data); err != nil {
		logger.MethodFailure(ctx, "template.Execute", err)
		http_helpers.WriteInternalServerError(ctx, w)
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
