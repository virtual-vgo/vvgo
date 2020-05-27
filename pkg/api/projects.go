package api

import (
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"net/http"
)

type ProjectsHandler struct{}

func (x ProjectsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	projectName := r.FormValue("name")
	project := projects.GetName(projectName)
	if project == nil {
		notFound(w)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	jsonEncode(w, project)
}
