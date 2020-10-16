package api

import (
	"bytes"
	"context"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
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

func identityFromContext(ctx context.Context) *login.Identity {
	ctxIdentity := ctx.Value(CtxKeyVVGOIdentity)
	identity, ok := ctxIdentity.(*login.Identity)
	if !ok {
		identity = new(login.Identity)
		*identity = login.Anonymous()
	}
	return identity
}

func parseAndExecute(ctx context.Context, dest io.Writer, data interface{}, templateFile string) bool {
	identity := identityFromContext(ctx)

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(map[string]interface{}{
		"link_to_template": func() string { return "https://github.com/virtual-vgo/vvgo/blob/master/public" + templateFile },
		"user_info":        identity.Info,
		"user_logged_in":   func() bool { return identity.IsAnonymous() },
		"user_is_member":   func() bool { return identity.HasRole(login.RoleVVGOMember) },
		"user_is_leader":   func() bool { return identity.HasRole(login.RoleVVGOLeader) },
		"user_on_teams":    func() bool { return identity.HasRole(login.RoleVVGOTeams) },
		"template_file":    func() string { return templateFile },
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
