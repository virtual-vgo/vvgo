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
	"net/url"
	"strings"
)

func Spreadsheet(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		return handleGetSpreadsheet(ctx, r.URL.Query())
	case http.MethodPost:
		return handlePostSpreadsheet(ctx, r.Body)
	default:
		return http_helpers.NewMethodNotAllowedError()
	}
}

type GetSpreadsheetRequest struct {
	SpreadsheetName string
	SheetNames      []string
}

func handleGetSpreadsheet(ctx context.Context, params url.Values) models.ApiResponse {
	data := GetSpreadsheetRequest{
		SpreadsheetName: params.Get("spreadsheetName"),
		SheetNames:      strings.Split(params.Get("sheetNames"), ","),
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

func handlePostSpreadsheet(ctx context.Context, body io.Reader) models.ApiResponse {
	var data models.Spreadsheet
	if err := json.NewDecoder(body).Decode(&data); err != nil {
		return http_helpers.NewJsonDecodeError(err)
	}

	for _, sheet := range data.Sheets {
		if err := redis.WriteSheet(ctx, data.SpreadsheetName, sheet.Name, sheet.Values); err != nil {
			logger.RedisFailure(ctx, err)
			return http_helpers.NewInternalServerError()
		}
	}
	return http_helpers.NewOkResponse()
}
