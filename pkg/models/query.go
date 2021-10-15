package models

import (
	"reflect"
	"strings"
)

type Query map[string]interface{}

func NewQuery() Query { return make(Query) }

func (x Query) WithField(key string, value interface{}) Query { x[key] = value; return x }
func (x Query) WithFieldsUnsafe(args ...interface{}) Query {
	for i := 0; i+1 < len(args); i += 2 {
		name, ok := args[i].(string)
		if !ok {
			panic("field names should be strings")
		}
		x[name] = args[i+1]
	}
	return x
}

// MatchSlice matches the slice of struct w/ the query.
// src should be a slice of structs.
// dest should be a pointer to a slice of structs and should have capacity to store matches.
// Matches are copied to dest.
func (x Query) MatchSlice(src, dest interface{}) {
	matches := 0
	srcVals := reflect.ValueOf(src)
	destVals := reflect.ValueOf(dest).Elem()
	for i := 0; i < srcVals.Len(); i++ {
		val := srcVals.Index(i)
		if x.MatchStruct(val.Interface()) {
			matches += 1
			if destVals.Len() < matches {
				destVals.SetLen(matches)
			}
			destVals.Index(matches - 1).Set(val)
		}
	}
	reflect.ValueOf(dest).Elem().Set(destVals)
}

func (x Query) MatchStruct(str interface{}) bool {
	wantRow := true
	reflectType := reflect.TypeOf(str)
	query := x.Query()
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		wantValue, ok := query[field.Name]
		if !ok {
			continue
		}
		if wantValue != reflect.ValueOf(str).Field(i).Interface() {
			wantRow = false
		}
	}
	return wantRow
}

func (x Query) Query() Query {
	clean := NewQuery()
	for k, v := range x {
		clean[strings.ReplaceAll(k, " ", "")] = v
	}
	return clean
}
