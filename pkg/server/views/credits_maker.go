package views

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"net/http"
)

func CreditsMaker(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data := struct {
		SpreadsheetID string
		ReadRange     string
		Project       string
		ErrorMessage  string
		WebsitePasta  string
		VideoPasta    string
		YoutubePasta  string
	}{
		SpreadsheetID: r.FormValue("spreadsheetID"),
		ReadRange:     r.FormValue("readRange"),
		Project:       r.FormValue("project"),
	}

	if data.SpreadsheetID != "" && data.ReadRange != "" {
		submissions, err := models.ListSubmissions(ctx, data.SpreadsheetID, data.ReadRange)
		if err != nil {
			logger.ListSubmissionsFailure(ctx, err)
			data.ErrorMessage = err.Error()
		} else {
			credits := submissions.ToCredits(data.Project)
			data.WebsitePasta = credits.WebsitePasta()
			data.VideoPasta = credits.VideoPasta()
			data.YoutubePasta = credits.YoutubePasta()
		}
	}

	ParseAndExecute(ctx, w, r, data, "credits-maker.gohtml")
}
