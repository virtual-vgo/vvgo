package api

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"strings"
)

func parseAndExecute(dest io.Writer, data interface{}, templateFiles ...string) bool {
	uploadTemplate, err := template.ParseFiles(templateFiles...)
	if err != nil {
		logger.WithError(err).Error("template.ParseFiles() failed")
		return false
	}
	if err := uploadTemplate.Execute(dest, &data); err != nil {
		logger.WithError(err).Error("template.Execute() failed")
		return false
	}
	return true
}

func jsonEncode(dest io.Writer, data interface{}) bool {
	if err := json.NewEncoder(dest).Encode(data); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		return false
	}
	return true
}

func acceptsType(r *http.Request, mimeType string) bool {
	for _, value := range r.Header["Accept"] {
		for _, wantType := range strings.Split(value, ",") {
			if wantType == mimeType {
				return true
			}
		}
	}
	return false
}
