package website_data

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
)

type Spreadsheet struct {
	SpreadsheetName string  `json:"spreadsheet_name"`
	Sheets          []Sheet `json:"sheets"`
}

func ServeSpreadsheet(r *http.Request) api.Response {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		return response.NewMethodNotAllowedError()
	}

	var data Spreadsheet
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return response.NewJsonDecodeError(err)
	}

	for _, sheet := range data.Sheets {
		if err := redis.WriteSheet(ctx, data.SpreadsheetName, sheet.Name, sheet.Values); err != nil {
			logger.RedisFailure(ctx, err)
			return response.NewInternalServerError()
		}
	}
	return api.NewOkResponse()
}
