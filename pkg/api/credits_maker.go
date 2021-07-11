package api

import (
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

type CreditsMakerConfig struct {
	DefaultSpreadsheetID string `redis:"default_spreadsheet_id"`
	DefaultReadRange     string `redis:"default_read_range"`
	DefaultProject       string `redis:"default_project"`
}

var CreditsMaker = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	var config CreditsMakerConfig
	_ = parse_config.ReadFromRedisHash(ctx, "credits_maker", &config)

	// set defaults
	if data.SpreadsheetID == "" {
		data.SpreadsheetID = "1a-2u726Hg-Wp5GMWfLnYwSi2DvTMym85gQqpRviafJk"
	}
	if data.ReadRange == "" {
		data.ReadRange = "Project 11: Prologue (Book One)!A3:I270"
	}
	if data.Project == "" {
		data.Project = "11-prologue-book-one"
	}
	ParseAndExecute(ctx, w, r, &data, "credits-maker.gohtml")
})
