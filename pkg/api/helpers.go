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

func badRequest(w http.ResponseWriter, reason string) {
	http.Error(w, reason, http.StatusBadRequest)
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "", http.StatusInternalServerError)
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "", http.StatusMethodNotAllowed)
}

func tooManyBytes(w http.ResponseWriter) {
	http.Error(w, "", http.StatusRequestEntityTooLarge)
}

func invalidContent(w http.ResponseWriter) {
	http.Error(w, "", http.StatusUnsupportedMediaType)
}
