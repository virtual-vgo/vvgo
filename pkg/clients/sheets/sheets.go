package sheets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"google.golang.org/api/sheets/v4"
	"reflect"
	"strconv"
)

const CacheTTL = "5"

func ReadSheet(ctx context.Context, spreadsheetID string, readRange string) ([][]interface{}, error) {
	values := ReadValuesFromRedis(ctx, spreadsheetID, readRange)
	if len(values) != 0 || ctx.Value("no_op_sheets") == true {
		return values, nil
	}

	values, err := readValuesFromSheets(ctx, spreadsheetID, readRange)
	if err != nil {
		logger.WithError(err).Errorf("failed to read spreadsheet values from sheets")
		return nil, err
	} else if len(values) != 0 {
		WriteValuesToRedis(ctx, spreadsheetID, readRange, values)
		return values, nil
	}
	return nil, fmt.Errorf("no data")
}

func readValuesFromSheets(ctx context.Context, spreadsheetID string, readRange string) ([][]interface{}, error) {
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Sheets client: %w", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data from sheet: %w", err)
	}
	return resp.Values, nil
}

func ReadValuesFromRedis(ctx context.Context, spreadsheetID string, readRange string) [][]interface{} {
	var buf bytes.Buffer
	key := "sheets:" + spreadsheetID + ":" + readRange
	if err := redis.Do(ctx, redis.Cmd(&buf, "GET", key)); err != nil {
		logger.WithError(err).Error("failed to read spreadsheet values from redis")
		return nil
	} else if buf.Len() == 0 {
		logger.WithField("key", key).Info("cache miss")
		return nil
	}

	var values [][]interface{}
	if err := json.NewDecoder(&buf).Decode(&values); err != nil {
		logger.JsonDecodeFailure(ctx, err)
		return nil
	}
	return values
}

func WriteValuesToSheets(ctx context.Context, spreadsheetID string, readRange string, values [][]interface{}) error {
	WriteValuesToRedis(ctx, spreadsheetID, readRange, values)

	if ctx.Value("no_op_sheets") == true {
		return nil
	}

	srv, err := sheets.NewService(ctx)
	if err != nil {
		return fmt.Errorf("sheets.NewService(): %w", err)
	}
	_, err = srv.Spreadsheets.Values.
		Update(spreadsheetID, readRange, &sheets.ValueRange{Values: values, MajorDimension: "ROWS"}).
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		return fmt.Errorf("sheets.Values.Update() failed: %w", err)
	}
	return nil
}

func WriteValuesToRedis(ctx context.Context, spreadsheetID string, readRange string, values [][]interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&values); err != nil {
		logger.JsonEncodeFailure(ctx, err)
		return
	}

	key := "sheets:" + spreadsheetID + ":" + readRange
	if err := redis.Do(ctx, redis.Cmd(nil, "SETEX", key, CacheTTL, buf.String())); err != nil {
		logger.WithError(err).Errorf("failed to write spreadsheet values to redis")
	}
}

func BuildIndex(fieldNames []interface{}) map[string]int {
	index := make(map[string]int, len(fieldNames))
	for i, col := range fieldNames {
		index[fmt.Sprintf("%s", col)] = i
	}
	return index
}

func ProcessRow(row []interface{}, dest interface{}, index map[string]int) {
	tagName := "col_name"
	if len(row) < 1 {
		return
	}
	reflectType := reflect.TypeOf(dest).Elem()
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		colName := field.Tag.Get(tagName)
		if colName == "" {
			colName = field.Name
		}
		colIndex, ok := index[colName]
		if !ok {
			continue
		}
		if len(row) > colIndex {
			switch field.Type.Kind() {
			case reflect.String:
				val := fmt.Sprint(row[colIndex])
				reflect.ValueOf(dest).Elem().Field(i).SetString(val)
			case reflect.Bool:
				val, _ := strconv.ParseBool(fmt.Sprint(row[colIndex]))
				reflect.ValueOf(dest).Elem().Field(i).SetBool(val)
			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				val, _ := strconv.ParseInt(fmt.Sprint(row[colIndex]), 10, 64)
				reflect.ValueOf(dest).Elem().Field(i).SetInt(val)
			}
		}
	}
}
