package api

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
)

type GetCreditsPastaRequest struct {
	SpreadsheetID string
	ReadRange     string
	ProjectName   string
}

func CreditsPasta(r *http.Request) models.ApiResponse {
	ctx := r.Context()

	inputData := GetCreditsPastaRequest{
		SpreadsheetID: r.FormValue("spreadsheetID"),
		ReadRange:     r.FormValue("readRange"),
		ProjectName:   r.FormValue("projectName"),
	}

	switch {
	case inputData.SpreadsheetID == "":
		return http_helpers.NewBadRequestError("spreadsheetID is required")
	case inputData.ReadRange == "":
		return http_helpers.NewBadRequestError("readRange is required")
	case inputData.ProjectName == "":
		return http_helpers.NewBadRequestError("projectName is required")
	default:
		break
	}

	submissions, err := models.ListSubmissions(ctx, inputData.SpreadsheetID, inputData.ReadRange)
	if err != nil {
		logger.ListSubmissionsFailure(ctx, err)
		return http_helpers.NewBadRequestError(err.Error())
	}

	credits := submissions.ToCredits(inputData.ProjectName)
	return models.ApiResponse{
		Status: models.StatusOk,
		CreditsPasta: &models.CreditsPasta{
			WebsitePasta: credits.WebsitePasta(),
			VideoPasta:   credits.VideoPasta(),
			YoutubePasta: credits.YoutubePasta(),
		},
	}
}
