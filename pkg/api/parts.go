package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

type PartView struct{}

func (x PartView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	ParseAndExecute(ctx, w, r, nil, "parts.gohtml")
}

type PartsAPI struct{}

func (x PartsAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	parts, err := sheets.ListParts(ctx, IdentityFromContext(ctx))
	if err != nil {
		logger.WithError(err).Error("valuesToProjects() failed")
		internalServerError(w)
		return
	}
	if project := r.FormValue("project"); project != "" {
		parts = parts.ForProject(project)
	}
	if parts == nil {
		parts = sheets.Parts{}
	}
	json.NewEncoder(w).Encode(parts)
}
