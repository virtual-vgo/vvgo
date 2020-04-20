package api

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type LoginHandler struct {
	NavBar   NavBar
	Secret   sessions.Secret
	Sessions *sessions.Store
}

func (x LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "login_handler")
	defer span.Send()

	switch r.Method {
	case http.MethodGet:
		var session sessions.Session
		if err := x.Sessions.ReadFromRequest(ctx, r, &session); err == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		opts := x.NavBar.NewOpts(r)
		page := struct {
			Header template.HTML
			NavBar template.HTML
		}{
			Header: Header(),
			NavBar: x.NavBar.RenderHTML(opts),
		}

		var buf bytes.Buffer
		if ok := parseAndExecute(&buf, &page, filepath.Join(PublicFiles, "login.gohtml")); !ok {
			internalServerError(w)
			return
		}
		buf.WriteTo(w)

	case http.MethodPost:
		var session sessions.Session
		if err := x.Sessions.ReadFromRequest(ctx, r, &session); err == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		user := r.FormValue("user")
		pass := r.FormValue("pass")
		if !(user == "jackson@jacksonargo.com" || pass == "jackson") {
			logger.WithFields(logrus.Fields{
				"user": user,
				"pass": pass,
			}).Error("authorization failed")
			unauthorized(w)
			return
		}

		cookie := x.Sessions.NewSessionCookie(time.Now().Add(7 * 24 * 3600 * time.Second))
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusFound)

	default:
		methodNotAllowed(w)
	}
}
