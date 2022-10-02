package api

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
)

func AllowedDatasets() []string {
	return []string{
		"Highlights",
		"Leaders",
		"Directors",
		"Roster",
		"Credits",
		"Instruments",
	}
}

func datasetIsAllowed(name string) bool {
	for _, dataset := range AllowedDatasets() {
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
