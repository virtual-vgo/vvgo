package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

var PartsView = ServeTemplate("parts.gohtml")

func PartsApi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	projects, err := sheets.ListProjects(ctx, IdentityFromContext(ctx))
	if err != nil {
		logger.WithError(err).Error("listProjects() failed")
		internalServerError(w)
		return
	}

	parts, err := sheets.ListParts(ctx)
	if err != nil {
		logger.WithError(err).Error("listParts() failed")
		internalServerError(w)
		return
	}
	parts = parts.ForProject(projects.Names()...)

	if project := r.FormValue("project"); project != "" {
		parts = parts.ForProject(project)
	}
	if parts == nil {
		parts = sheets.Parts{}
	}
	json.NewEncoder(w).Encode(parts.Sort())
}
