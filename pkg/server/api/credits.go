package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Credits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectName := r.FormValue("project")
	if projectName == "" {
		http_helpers.BadRequest(ctx, w, "project is required")
		return
	}

	projects, err := models.ListProjects(ctx, login.IdentityFromContext(ctx))
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
		return
	}

	project, ok := projects.Get(projectName)
	if !ok {
		http_helpers.BadRequest(ctx, w, "requested project does not exist")
		return
	}

	credits, err := models.ListCredits(ctx)
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("HtmlSource-Type", "application/json")
	data := models.BuildCreditsTable(credits, project)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.JsonEncodeFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
	}
}
