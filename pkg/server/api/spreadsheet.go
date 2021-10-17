package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
)

func Spreadsheet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data models.Spreadsheet
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		http_helpers.BadRequest(ctx, w, "invalid json")
		return
	}

	switch r.Method {
	case http.MethodGet:
		for i := range data.Sheets {
			values, err := redis.ReadSheet(ctx, data.SpreadsheetName, data.Sheets[i].Name)
			if err != nil {
				logger.RedisFailure(ctx, err)
				http_helpers.InternalServerError(ctx, w)
				return
			}
			data.Sheets[i].Values = values
		}
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.JsonEncodeFailure(ctx, err)
			http_helpers.InternalServerError(ctx, w)
			return
		}
		http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
			Status: models.StatusOk,
		})
		return

	case http.MethodPost:
		for _, sheet := range data.Sheets {
			if err := redis.WriteSheet(ctx, data.SpreadsheetName, sheet.Name, sheet.Values); err != nil {
				logger.RedisFailure(ctx, err)
				http_helpers.InternalServerError(ctx, w)
				return
			}
		}
		return

	default:
		http_helpers.MethodNotAllowed(ctx, w)
		return
	}
}
