package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"net/http"
)

type SpreadsheetData struct {
	SpreadsheetName string `json:"spreadsheet_name"`

	Sheets []struct {
		Name   string          `json:"name"`
		Values [][]interface{} `json:"values"`
	} `json:"sheets"`
}

func Spreadsheet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var data SpreadsheetData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		helpers.BadRequest(w, "invalid json")
		return
	}

	switch r.Method {
	case http.MethodGet:
		for i := range data.Sheets {
			values, err := redis.ReadSheet(ctx, data.SpreadsheetName, data.Sheets[i].Name)
			if err != nil {
				logger.RedisFailure(ctx, err)
				helpers.InternalServerError(w)
				return
			}
			data.Sheets[i].Values = values
		}
		if err := json.NewEncoder(w).Encode(data); err != nil {
			logger.JsonEncodeFailure(ctx, err)
			helpers.InternalServerError(w)
			return
		}
		return

	case http.MethodPost:
		for _, sheet := range data.Sheets {
			if err := redis.WriteSheet(ctx, data.SpreadsheetName, sheet.Name, sheet.Values); err != nil {
				logger.RedisFailure(ctx, err)
				helpers.InternalServerError(w)
				return
			}
		}
		return

	default:
		helpers.MethodNotAllowed(w)
		return
	}
}
