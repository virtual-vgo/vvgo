package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Credits(w http.ResponseWriter, r *http.Request) {
	projectName := r.FormValue("project")
	if projectName == "" {
		helpers.BadRequest(w, "project is required")
		return
	}

	ctx := r.Context()
	projects, err := models.ListProjects(ctx, login.IdentityFromContext(ctx))
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		helpers.InternalServerError(w)
		return
	}

	project, ok := projects.Get(projectName)
	if !ok {
		helpers.BadRequest(w, "requested project does not exist")
		return
	}

	credits, err := models.ListCredits(ctx)
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		helpers.InternalServerError(w)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	data := models.BuildCreditsTable(credits, project)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.JsonEncodeFailure(ctx, err)
	}
}
