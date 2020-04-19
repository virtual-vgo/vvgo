package api

import (
	"bytes"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type LoginHandler struct {
	NavBar
}

func (x LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "login_handler")
	defer span.Send()

	var session sessions.Session
	if err := sessions.ReadFromRequest(ctx, r, &session); err != nil {
		user := r.FormValue("user")
		pass := r.FormValue("pass")
		if user == "jackson" && pass == "jackson" {
			cookie := http.Cookie{
				Name:    sessions.SessionKey,
				Value:   user,
				Expires: time.Now().Add(3600 * time.Second),
			}
			http.SetCookie(w, &cookie)
			sessions.Add(ctx, &sessions.Session{
				Key:       user,
				VVVGOUser: user,
			})
			w.Write([]byte("welcome jackson, have a cookie!\n"))
		} else {
			unauthorized(w)
		}
	} else {
		fmt.Fprint(w, "welcome back "+session.VVVGOUser)
	}
}

func (x LoginHandler) __ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
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
		} else {
			buf.WriteTo(w)
		}
	case http.MethodPost:
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
