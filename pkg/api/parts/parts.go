package parts

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api/helpers"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

var logger = log.New()

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	projects, err := sheets.ListProjects(ctx, login.IdentityFromContext(ctx))
	if err != nil {
		logger.WithError(err).Error("listProjects() failed")
		helpers.InternalServerError(w)
		return
	}

	parts, err := sheets.ListParts(ctx)
	if err != nil {
		logger.WithError(err).Error("listParts() failed")
		helpers.InternalServerError(w)
		return
	}
	parts = parts.ForProject(projects.Names()...)

	if project := r.FormValue("project"); project != "" {
		parts = parts.ForProject(project)
	}
	if parts == nil {
		parts = sheets.Parts{}
	}

	if err := json.NewEncoder(w).Encode(parts.Sort()); err != nil {
		logger.JsonEncodeFailure(ctx, err)
	}
}
