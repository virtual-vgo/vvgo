package vvgo

import (
	"bytes"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/errors"
	"github.com/virtual-vgo/vvgo/pkg/http_wrappers"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"net/http"
)

const Endpoint = "https://vvgo.org/api/v1"

func GetSheets(spreadsheet string, sheets ...string) (models.Spreadsheet, error) {
	data := models.Spreadsheet{
		SpreadsheetName: spreadsheet,
	}
	for _, sheet := range sheets {
		data.Sheets = append(data.Sheets, models.Sheet{Name: sheet})
	}

	req, err := NewRequest(http.MethodGet, Endpoint+"/spreadsheet", data)
	if err != nil {
		return models.Spreadsheet{}, errors.NewRequestFailure(err)
	}

	if err := DoRequest(req, &data); err != nil {
		return models.Spreadsheet{}, err
	}
	return data, nil
}

func DoRequest(req *http.Request, dest interface{}) error {
	resp, err := http_wrappers.DoRequest(req)
	if err != nil {
		return errors.HttpDoFailure(err)
	}

	if resp.StatusCode != 200 {
		return errors.Non200StatusCode()
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return errors.JsonDecodeFailure(err)
	}
	return nil
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
