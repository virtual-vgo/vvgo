package website_data

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Sheet struct {
	Name   string
	Values [][]interface{}
}

func UnmarshalSheet(rows [][]interface{}, dest interface{}) {
	colNames := make([]string, len(rows[0]))
	for i := range colNames {
		colNames[i] = fmt.Sprintf("%s", rows[0][i])
		colNames[i] = strings.ReplaceAll(colNames[i], " ", "")
	}

	destValue := reflect.ValueOf(dest).Elem()
	dataType := destValue.Type().Elem()
	for _, row := range rows[1:] {
		data := reflect.New(dataType)
		for i := range row {
			field := data.Elem().FieldByName(colNames[i])
			if field.Kind() == reflect.Invalid {
				continue
			}
			switch field.Type().Kind() {
			case reflect.String:
				field.SetString(fmt.Sprint(row[i]))
			case reflect.Bool:
				val, _ := strconv.ParseBool(fmt.Sprint(row[i]))
				field.SetBool(val)
			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				val, _ := strconv.ParseInt(fmt.Sprint(row[i]), 10, 64)
				field.SetInt(val)
			}
		}
		destValue.Set(reflect.Append(destValue, data.Elem()))
	}
	reflect.ValueOf(dest).Elem().Set(destValue)
}
