package website_data

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"strings"
)

func AllowedDatasets() []string {
	return []string{
		"Highlights",
		"Leaders",
		"Directors",
		"Roster",
		"Credits",
	}
}

func datasetIsAllowed(name string) bool {
	for _, dataset := range AllowedDatasets() {
		if name == dataset {
			return true
		}
	}
	return false
}

type GetDatasetUrlParams struct{ Name string }

func ServeDataset(r *http.Request) api.Response {
	ctx := r.Context()
	var dataset GetDatasetUrlParams
	dataset.Name = r.URL.Query().Get("name")
	switch {
	case dataset.Name == "":
		return response.NewBadRequestError("name cannot be empty")
	case datasetIsAllowed(dataset.Name) == false:
		return response.NewErrorResponse(response.Error{
			Code:    http.StatusForbidden,
			Message: fmt.Sprintf("sheet `%s` is not allowed", dataset.Name),
		})
	default:
		sheetData, err := redis.ReadSheet(ctx, SpreadsheetWebsiteData, dataset.Name)
		if err != nil {
			logger.RedisFailure(ctx, err)
			return response.NewInternalServerError()
		}
		return api.Response{
			Status:  api.StatusOk,
			Dataset: valuesToMap(sheetData),
		}
	}
}

func valuesToMap(rows [][]interface{}) []map[string]string {
	if len(rows) == 0 {
		return nil
	}

	colNames := make([]string, len(rows[0]))
	for i := range colNames {
		colNames[i] = fmt.Sprintf("%s", rows[0][i])
		colNames[i] = strings.ReplaceAll(colNames[i], " ", "")
	}
	data := make([]map[string]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		rowMap := make(map[string]string, len(colNames))
		for j := range colNames {
			if j < len(row) {
				rowMap[colNames[j]] = fmt.Sprintf("%s", row[j])
			}
		}
		data = append(data, rowMap)
	}
	return data
}
