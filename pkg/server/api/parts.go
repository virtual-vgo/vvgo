package api

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Parts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	projects, err := models.ListProjects(ctx, login.IdentityFromContext(ctx))
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
		return
	}

	parts, err := models.ListParts(ctx)
	if err != nil {
		logger.ListPartsFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
		return
	}
	parts = parts.ForProject(projects.Names()...)

	if project := r.FormValue("project"); project != "" {
		parts = parts.ForProject(project)
	}
	if parts == nil {
		parts = models.Parts{}
	}

	http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
		Status: models.StatusOk,
		Type:   models.ResponseTypeParts,
		Parts:  &models.PartsResponse{Parts: parts.Sort()},
	})
}
