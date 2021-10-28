package api

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Projects(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	projects, err := models.ListProjects(ctx, login.IdentityFromContext(ctx))
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		return http_helpers.NewInternalServerError()
	}

	if projects == nil {
		projects = []models.Project{}
	}
	return models.ApiResponse{Status: models.StatusOk, Projects: projects.ReverseSort()}
}
