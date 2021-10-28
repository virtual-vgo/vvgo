package api

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

type GetCreditsTableRequest struct {
	ProjectName string
}

func CreditsTable(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)

	var data GetCreditsTableRequest
	data.ProjectName = r.URL.Query().Get("projectName")
	if data.ProjectName == "" {
		return http_helpers.NewBadRequestError("projectName is required")
	}

	projects, err := models.ListProjects(ctx, identity)
	if err != nil {
		logger.ListProjectsFailure(ctx, err)
		return http_helpers.NewInternalServerError()
	}
	wantProject, ok := projects.Get(data.ProjectName)
	if !ok {
		return http_helpers.NewNotFoundError(fmt.Sprintf("project %s not found", data.ProjectName))
	}

	credits, err := models.ListCredits(ctx)
	if err != nil {
		logger.ListCreditsFailure(ctx, err)
		return http_helpers.NewInternalServerError()
	}

	return models.ApiResponse{Status: models.StatusOk, CreditsTable: models.BuildCreditsTable(credits, wantProject)}
}
