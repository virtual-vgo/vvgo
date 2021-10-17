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
	}
	for _, dataset := range allowedDatasets {
		if name == dataset {
			return true
		}
	}
	return false
}

type DatasetRequest struct{ Name string }

func Dataset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var dataset DatasetRequest
	dataset.Name = r.URL.Query().Get("name")
	switch {
	case dataset.Name == "":
		http_helpers.WriteErrorResponse(ctx, w, models.ErrorResponse{
			Code:  http.StatusBadRequest,
			Error: "name cannot be empty",
		})
	case datasetIsAllowed(dataset.Name) == false:
		http_helpers.WriteErrorResponse(ctx, w, models.ErrorResponse{
			Code:  http.StatusForbidden,
			Error: fmt.Sprintf("sheet `%s` is not allowed", dataset.Name),
		})
	default:
		sheetData, err := redis.ReadSheet(ctx, models.SpreadsheetWebsiteData, dataset.Name)
		if err != nil {
			logger.RedisFailure(ctx, err)
			http_helpers.InternalServerError(ctx, w)
			return
		}
		http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
			Status:  models.StatusOk,
			Dataset: &models.Dataset{Name: dataset.Name, Rows: models.ValuesToMap(sheetData)},
		})
	}
}

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

func Parts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)

	parts, err := models.ListParts(ctx, identity)
	if err != nil {
		logger.ListPartsFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
		return
	}

	if parts == nil {
		parts = []models.Part{}
	}
	parts = parts.Sort()
	http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
		Status: models.StatusOk,
		Parts:  parts,
	})
}
