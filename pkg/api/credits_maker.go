package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

type CreditsMaker struct{ Template }

func (x CreditsMaker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		submissions, err := sheets.ListSubmissions(ctx, data.SpreadsheetID, data.ReadRange)
		if err != nil {
			logger.WithError(err).Error("readSheet() failed")
			data.ErrorMessage = err.Error()
		} else {
			credits := submissions.ToCredits(data.Project)
			data.WebsitePasta = credits.WebsitePasta()
			data.VideoPasta = credits.VideoPasta()
			data.YoutubePasta = credits.YoutubePasta()
		}
	}

	// set defaults
	if data.SpreadsheetID == "" {
		data.SpreadsheetID = "1BP3fGC2C6mKe3ZuVhby4eCxidlHL768bDdHsJ5mQleo"
	}
	if data.ReadRange == "" {
		data.ReadRange = "06 Aurene!A3:I39"
	}
	if data.Project == "" {
		data.Project = "06-aurene-dragon-full-of-light"
	}
	x.Template.ParseAndExecute(ctx, w, r, &data, "credits-maker.gohtml")
}
