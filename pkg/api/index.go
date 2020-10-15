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

	page := struct {
		NavBar NavBarOpts
	}{
		NavBar: NewNavBarOpts(ctx),
	}

	var buffer bytes.Buffer
	if ok := parseAndExecute(ctx, &buffer, &page, PublicFiles+"/index.gohtml"); !ok {
		internalServerError(w)
		return
	}
	buffer.WriteTo(w)
}

type NavBarOpts struct {
	ShowLogin       bool
	ShowMemberLinks bool
	ShowTeamsLinks  bool
	PartsActive     bool
	LoginActive     bool
	ProjectsActive  bool
}

func NewNavBarOpts(ctx context.Context) NavBarOpts {
	identity := identityFromContext(ctx)
	return NavBarOpts{
		ShowMemberLinks: identity.HasRole(login.RoleVVGOMember),
		ShowTeamsLinks:  identity.HasRole(login.RoleVVGOTeams),
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

func parseAndExecute(ctx context.Context, dest io.Writer, data interface{}, templateFile string) bool {
	identity := identityFromContext(ctx)

	tmpl, err := template.New(filepath.Base(templateFile)).Funcs(map[string]interface{}{
		"link_to_template": func() string { return "https://github.com/virtual-vgo/vvgo/blob/master/" + templateFile },
		"user_info":        identity.Info,
		"user_is_leader":   func() bool { return identity.HasRole(login.RoleVVGOLeader) },
		"user_on_teams":    func() bool { return identity.HasRole(login.RoleVVGOTeams) },
	}).ParseFiles(
		templateFile,
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
