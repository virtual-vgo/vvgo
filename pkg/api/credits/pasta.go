package credits

import (
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
)

type GetPastaFormParams struct {
	SpreadsheetID string
	ReadRange     string
	ProjectName   string
}

type Pasta struct {
	WebsitePasta string
	VideoPasta   string
	YoutubePasta string
}

func ServePasta(r *http.Request) api.Response {
	ctx := r.Context()

	inputData := GetPastaFormParams{
		SpreadsheetID: r.FormValue("spreadsheetID"),
		ReadRange:     r.FormValue("readRange"),
		ProjectName:   r.FormValue("projectName"),
	}

	switch {
	case inputData.SpreadsheetID == "":
		return response.NewBadRequestError("spreadsheetID is required")
	case inputData.ReadRange == "":
		return response.NewBadRequestError("readRange is required")
	case inputData.ProjectName == "":
		return response.NewBadRequestError("projectName is required")
	default:
		break
	}

	submissions, err := ListSubmissions(ctx, inputData.SpreadsheetID, inputData.ReadRange)
	if err != nil {
		logger.ListSubmissionsFailure(ctx, err)
		return response.NewBadRequestError(err.Error())
	}

	credits := submissions.ToCredits(inputData.ProjectName)
	return api.Response{
		Status: api.StatusOk,
		CreditsPasta: &Pasta{
			WebsitePasta: credits.WebsitePasta(),
			VideoPasta:   credits.VideoPasta(),
			YoutubePasta: credits.YoutubePasta(),
		},
	}
}
