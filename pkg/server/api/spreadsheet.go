package api

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"io"
	"net/http"
)

func Spreadsheet(r *http.Request) models.ApiResponse {
	ctx := r.Context()

	var data models.Spreadsheet
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return http_helpers.NewJsonDecodeError(err)
	}

	switch r.Method {
	case http.MethodGet:
		return handleGetSpreadsheet(ctx, r.Body)
	case http.MethodPost:
		return handlePostSpreadsheet(data, ctx)
	default:
		return http_helpers.NewMethodNotAllowedError()
	}
}

type GetSpreadsheetRequest struct {
	SpreadsheetName string
	SheetNames      []string
}

func handleGetSpreadsheet(ctx context.Context, body io.Reader) models.ApiResponse {
	var data GetSpreadsheetRequest
	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return http_helpers.NewJsonDecodeError(err)
	}

	var sheets []models.Sheet
	for _, sheetName := range data.SheetNames {
		values, err := redis.ReadSheet(ctx, data.SpreadsheetName, sheetName)
		if err != nil {
			logger.RedisFailure(ctx, err)
			return http_helpers.NewInternalServerError()
		}
		sheets = append(sheets, models.Sheet{Name: sheetName, Values: values})
	}
	return models.ApiResponse{Status: models.StatusOk, Spreadsheet: &models.Spreadsheet{
		SpreadsheetName: data.SpreadsheetName,
		Sheets:          sheets,
	}}
}

func handlePostSpreadsheet(data models.Spreadsheet, ctx context.Context) models.ApiResponse {
	for _, sheet := range data.Sheets {
		if err := redis.WriteSheet(ctx, data.SpreadsheetName, sheet.Name, sheet.Values); err != nil {
			logger.RedisFailure(ctx, err)
			return http_helpers.NewInternalServerError()
		}
	}
	return http_helpers.NewOkResponse()
}
