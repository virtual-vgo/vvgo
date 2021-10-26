package api

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func datasetIsAllowed(name string) bool {
	allowedDatasets := []string{
		"Highlights",
		"Leaders",
		"Directors",
		"Roster",
		"Credits",
	}
	for _, dataset := range allowedDatasets {
		if name == dataset {
			return true
		}
	}
	return false
}

type DatasetRequest struct{ Name string }

func Dataset(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	var dataset DatasetRequest
	dataset.Name = r.URL.Query().Get("name")
	switch {
	case dataset.Name == "":
		return http_helpers.NewBadRequestError("name cannot be empty")
	case datasetIsAllowed(dataset.Name) == false:
		return http_helpers.NewErrorResponse(models.ApiError{
			Code:  http.StatusForbidden,
			Error: fmt.Sprintf("sheet `%s` is not allowed", dataset.Name),
		})
	default:
		sheetData, err := redis.ReadSheet(ctx, models.SpreadsheetWebsiteData, dataset.Name)
		if err != nil {
			logger.RedisFailure(ctx, err)
			return http_helpers.NewInternalServerError()
		}
		return models.ApiResponse{
			Status:  models.StatusOk,
			Dataset: models.ValuesToMap(sheetData),
		}
	}
}

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
	return models.ApiResponse{Status: models.StatusOk, Projects: projects.Sort()}
}

func Parts(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)

	parts, err := models.ListParts(ctx, identity)
	if err != nil {
		logger.ListPartsFailure(ctx, err)
		return http_helpers.NewInternalServerError()
	}

	if parts == nil {
		parts = []models.Part{}
	}
	return models.ApiResponse{Status: models.StatusOk, Parts: parts.Sort()}
}
