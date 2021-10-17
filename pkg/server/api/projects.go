package api

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Projects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	projects, err := models.ListProjects(ctx, login.IdentityFromContext(ctx))
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
		return
	}

	if r.FormValue("latest") == "true" {
		project := projects.WithField("Video Released", true).Sort().Last()
		projects = models.Projects{project}
	}

	http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
		Status:   models.StatusOk,
		Type:     models.ResponseTypeProjects,
		Projects: &models.ProjectsResponse{Projects: projects.Sort()},
	})
}
