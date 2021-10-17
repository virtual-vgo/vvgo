package models

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"reflect"
)

const SheetDirectors = "Leaders"

type Directors []Director

type Director struct {
	Name         string
	Epithet      string
	Affiliations string
	Blurb        string
	Icon         string
}

func ListDirectors(ctx context.Context) (Directors, error) {
	values, err := redis.ReadSheet(ctx, SpreadsheetWebsiteData, SheetDirectors)
	if err != nil {
		return nil, err
	}
	return valuesToDirectors(values), nil
}

func valuesToDirectors(values [][]interface{}) Directors {
	if len(values) < 1 {
		return nil
	}
	leaders := make([]Director, 0, len(values)-1)
	UnmarshalSheet(values, &leaders)
	return leaders
}

func directorsToValues(leaders Directors) [][]interface{} {
	values := make([][]interface{}, 1, len(leaders)+1)
	values[0] = structToColNames(Director{})
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
