package models

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/clients/sheets"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"reflect"
)

type Leaders []Leader

type Leader struct {
	Name         string
	Epithet      string
	Affiliations string
	Blurb        string
	Icon         string
}

func ListLeaders(ctx context.Context) (Leaders, error) {
	values, err := sheets.ReadSheet(ctx, parse_config.Config.Sheets.WebsiteDataSpreadsheetID, "Leaders")
	if err != nil {
		return nil, err
	}
	return valuesToLeaders(values), nil
}

func valuesToLeaders(values [][]interface{}) Leaders {
	if len(values) < 1 {
		return nil
	}
	index := sheets.BuildIndex(values[0])
	leaders := make([]Leader, len(values)-1)
	for i, row := range values[1:] {
		sheets.ProcessRow(row, &leaders[i], index)
	}
	return leaders
}

func leadersToValues(leaders Leaders) [][]interface{} {
	values := make([][]interface{}, 1, len(leaders)+1)
	values[0] = structToColNames(Leader{})
	for _, leader := range leaders {
		values = append(values, structToValueRow(leader))
	}
	return values
}

func structToColNames(str interface{}) []interface{} {
	var colNames []interface{}
	tagName := "col_name"
	reflectType := reflect.TypeOf(str)
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		colName := field.Tag.Get(tagName)
		if colName == "" {
			colName = field.Name
		}
		colNames = append(colNames, colName)
	}
	return colNames
}

func structToValueRow(str interface{}) []interface{} {
	var values []interface{}
	strValue := reflect.ValueOf(str)
	for i := 0; i < strValue.NumField(); i++ {
		values = append(values, strValue.Field(i).Interface())
	}
	return values
}