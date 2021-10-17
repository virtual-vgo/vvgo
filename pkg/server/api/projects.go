package api

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Projects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projects, err := models.ListProjects(ctx, login.IdentityFromContext(ctx))
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
		return
	}

	if projects == nil {
		projects = []models.Project{}
	}
	projects = projects.Sort()
	http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
		Status:   models.StatusOk,
		Projects: projects,
	})
}
