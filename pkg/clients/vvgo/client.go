package vvgo

import (
	"bytes"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/website_data"
	"github.com/virtual-vgo/vvgo/pkg/clients/http_util"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/errors"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
)

const Endpoint = "https://vvgo.org/api/v1"

func GetSheets(spreadsheet string, sheets ...string) (website_data.Spreadsheet, error) {
	req, err := NewRequest(http.MethodGet, Endpoint+"/spreadsheet",
		&website_data.GetSpreadsheetRequest{SpreadsheetName: spreadsheet, SheetNames: sheets})
	if err != nil {
		return website_data.Spreadsheet{}, errors.NewRequestFailure(err)
	}

	data, err := DoRequest(req)

	switch {
	case err != nil:
		return website_data.Spreadsheet{}, err
	case data.Error != nil:
		return website_data.Spreadsheet{}, errors.New(data.Error.Message)
	case data.Spreadsheet == nil:
		return website_data.Spreadsheet{}, errors.New("invalid api response data")
	default:
		return *data.Spreadsheet, nil
	}
}

func DoRequest(r *http.Request) (api.Response, error) {
	ctx := r.Context()
	resp, err := http_util.DoRequest(r)
	if err != nil {
		return api.Response{}, errors.HttpDoFailure(err)
	}

	var data api.Response
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		return api.Response{}, errors.New("invalid response from api")
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
	req.Header.Set("Authorization", "Bearer "+config.Env.VVGO.ClientToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "vvgo-client")
	return req, nil
}
