package models

import "reflect"

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
}

func (x Query) MatchStruct(str interface{}) bool {
	tagName := "col_name"
	wantRow := true
	reflectType := reflect.TypeOf(str)
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		colName := field.Tag.Get(tagName)
		if colName == "" {
			colName = field.Name
		}
		wantValue, ok := x[colName]
		if !ok {
			continue
		}
		if wantValue != reflect.ValueOf(str).Field(i).Interface() {
			wantRow = false
		}
	}
	return wantRow
}
