package api

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Credits(r *http.Request) models.ApiResponse {
	ctx := r.Context()

	projectName := r.FormValue("project")
	if projectName == "" {
		return http_helpers.NewBadRequestError("project is requited")
	}

	projects, err := models.ListProjects(ctx, login.IdentityFromContext(ctx))
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		return http_helpers.NewInternalServerError()
	}

	project, ok := projects.Get(projectName)
	if !ok {
		return http_helpers.NewNotFoundError(fmt.Sprintf("project %s does not exist", projectName))
	}

	credits, err := models.ListCredits(ctx)
	if err != nil {
		logger.ListCreditsFailure(ctx, err)
		return http_helpers.NewInternalServerError()
	}

	data := models.BuildCreditsTable(credits, project)
	return models.ApiResponse{Status: models.StatusOk, CreditsTable: data}
}
