package api

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

type LoginHandler struct {
	NavBar
}

func (x LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, ok := sessions.Read(r)
	if !ok {
		user := r.FormValue("user")
		pass := r.FormValue("pass")
		if user == "jackson" && pass == "jackson" {
			cookie := http.Cookie{
				Name:    SessionCookie,
				Value:   user,
				Expires: time.Now().Add(3600 * time.Second),
			}
			http.SetCookie(w, &cookie)
			sessions.Add(Session{
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
