package vvgo

import (
	"bytes"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/errors"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/api"
	"net/http"
)

const Endpoint = "https://vvgo.org/api/v1"

func GetSheets(spreadsheet string, sheets ...string) (models.Spreadsheet, error) {
	req, err := NewRequest(http.MethodGet, Endpoint+"/spreadsheet",
		&api.GetSpreadsheetRequest{SpreadsheetName: spreadsheet, SheetNames: sheets})
	if err != nil {
		return models.Spreadsheet{}, errors.NewRequestFailure(err)
	}

	data, err := DoRequest(req)

	switch {
	case err != nil:
		return models.Spreadsheet{}, err
	case data.Error != nil:
		return models.Spreadsheet{}, errors.New(data.Error.Error)
	case data.Spreadsheet == nil:
		return models.Spreadsheet{}, errors.New("invalid api response data")
	default:
		return *data.Spreadsheet, nil
	}
}

func DoRequest(r *http.Request) (models.ApiResponse, error) {
	ctx := r.Context()
	resp, err := http_wrappers.DoRequest(r)
	if err != nil {
		return models.ApiResponse{}, errors.HttpDoFailure(err)
	}

	var data models.ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		return models.ApiResponse{}, errors.New("invalid response from api")
	}
	return data, nil
}

func NewRequest(method, url string, body interface{}) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		err := json.NewEncoder(&buf).Encode(body)
		if err != nil {
			return nil, errors.JsonEncodeFailure(err)
		}
	}

	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return nil, errors.NewRequestFailure(err)
	}
	req.Header.Set("Authorization", "Bearer "+config.Config.VVGO.ClientToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "vvgo-client")
	return req, nil
}
