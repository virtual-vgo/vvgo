package devel

import (
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/clients/vvgo"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
)

// Endpoints for devel tools

func FetchSpreadsheets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	spreadsheet, err := vvgo.GetSheets(models.SpreadsheetWebsiteData,
		models.SheetCredits, models.SheetProjects, models.SheetParts, models.SheetDirectors, "Highlights", "Roster", "Instruments")
	if err != nil {
		logger.MethodFailure(ctx, "vvgo.GetSheets", err)
		http_helpers.WriteInternalServerError(ctx, w)
		return
	}

	for _, sheet := range spreadsheet.Sheets {
		if err := redis.WriteSheet(ctx, spreadsheet.SpreadsheetName, sheet.Name, sheet.Values); err != nil {
			logger.RedisFailure(ctx, err)
			http_helpers.WriteInternalServerError(ctx, w)
			return
		}
	}
	http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
		Status:      models.StatusOk,
		Spreadsheet: &spreadsheet,
	})
}
