package api

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type LoginHandler struct {
	NavBar
}

func (x LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("vvgo_session")
	if err != nil {
		logger.WithError(err).Info("cookie error")
	}
	if err == nil && cookie.Name == "vvgo_session" && cookie.Value == "jackson" {
		w.Write([]byte("welcome back jackson\n"))
		return
	}

	user := r.FormValue("user")
	pass := r.FormValue("pass")
	if user == "jackson" && pass == "jackson" {
		http.SetCookie(w, &http.Cookie{
			Name:    "vvgo_session",
			Value:   "jackson",
			Expires: time.Now().Add(3600 * time.Second),
		})
		w.Write([]byte("welcome jackson, have a cookie!\n"))
	} else {
		unauthorized(w)
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
