package website_data

import (
	"encoding/json"
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
)

type Spreadsheet struct {
	SpreadsheetName string  `json:"spreadsheet_name"`
	Sheets          []Sheet `json:"sheets"`
}

func ServeSpreadsheet(r *http.Request) http2.Response {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		return errors.NewMethodNotAllowedError()
	}

	var data Spreadsheet
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return errors.NewJsonDecodeError(err)
	}

	for _, sheet := range data.Sheets {
		if err := redis.WriteSheet(ctx, data.SpreadsheetName, sheet.Name, sheet.Values); err != nil {
			logger.RedisFailure(ctx, err)
			return errors.NewInternalServerError()
		}
	}
	return http2.NewOkResponse()
}
