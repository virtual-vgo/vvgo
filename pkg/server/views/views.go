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
	"io/ioutil"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
)

var PublicFiles = "public"

var (
	Index           = ServeJsPage("Virtual Video Game Orchestra", "dist/index.js")
	ServePublicFile = http.FileServer(http.Dir(PublicFiles)).ServeHTTP
	Parts           = ServeTemplate("parts.gohtml")
	About           = ServeJsPage("About", "dist/about.js")
	ContactUs       = ServeHtml("Contact Us", "contact_us.html")
	Voting          = ServeTemplate("voting.gohtml")
	Sessions        = ServeJsPage("Manage Sessions", "dist/sessions.js")
)

type Page struct {
	Title    string
	JsSource string
	Content  template.HTML
}

func ServeHtml(title, htmlSource string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		content, err := ioutil.ReadFile(PublicFiles + "/" + htmlSource)
		if err != nil {
			logger.MethodFailure(ctx, "file.Read", err)
			http_helpers.InternalServerError(ctx, w)
			return
		}
		Page{Title: "VVGO | " + title, Content: template.HTML(content)}.Render(w, r)
	}
}

func ServeJsPage(title, jsSource string) http.HandlerFunc {
	return Page{Title: "VVGO | " + title, JsSource: jsSource}.Render
}

func (x Page) Render(w http.ResponseWriter, r *http.Request) {
	ParseAndExecute(r.Context(), w, r, x, "page.gohtml")
}

func ServeTemplate(templateFile string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ParseAndExecute(r.Context(), w, r, nil, templateFile)
	}
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
		"user_is_member":   func() bool { return identity.HasRole(models.RoleVVGOMember) },
		"user_is_leader":   func() bool { return identity.HasRole(models.RoleVVGOLeader) },
		"user_on_teams":    func() bool { return identity.HasRole(models.RoleVVGOTeams) },
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
		http_helpers.InternalServerError(ctx, w)
		return
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, &data); err != nil {
		logger.MethodFailure(ctx, "template.Execute", err)
		http_helpers.InternalServerError(ctx, w)
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
