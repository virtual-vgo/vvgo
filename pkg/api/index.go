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
