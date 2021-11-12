package credits

import (
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
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

func ServePasta(r *http.Request) http2.Response {
	ctx := r.Context()

	inputData := GetPastaFormParams{
		SpreadsheetID: r.FormValue("spreadsheetID"),
		ReadRange:     r.FormValue("readRange"),
		ProjectName:   r.FormValue("projectName"),
	}

	switch {
	case inputData.SpreadsheetID == "":
		return errors.NewBadRequestError("spreadsheetID is required")
	case inputData.ReadRange == "":
		return errors.NewBadRequestError("readRange is required")
	case inputData.ProjectName == "":
		return errors.NewBadRequestError("projectName is required")
	default:
		break
	}

	submissions, err := ListSubmissions(ctx, inputData.SpreadsheetID, inputData.ReadRange)
	if err != nil {
		logger.ListSubmissionsFailure(ctx, err)
		return errors.NewBadRequestError(err.Error())
	}

	credits := submissions.ToCredits(inputData.ProjectName)
	return http2.Response{
		Status: http2.StatusOk,
		CreditsPasta: &Pasta{
			WebsitePasta: credits.WebsitePasta(),
			VideoPasta:   credits.VideoPasta(),
			YoutubePasta: credits.YoutubePasta(),
		},
	}
}
