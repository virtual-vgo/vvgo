package api

import (
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"strings"
)

func readBody(dest io.Writer, r *http.Request) bool {
	switch r.Header.Get("Content-Encoding") {
	case "application/gzip":
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			logger.WithError(err).Error("gzip.NewReader() failed")
			return false
		}
		if _, err := io.Copy(dest, gzipReader); err != nil {
			logger.WithError(err).Error("gzipReader.Read() failed")
			return false
		}
		if err := gzipReader.Close(); err != nil {
			logger.WithError(err).Error("gzipReader.Close() failed")
			return false
		}

	default:
		if _, err := io.Copy(dest, r.Body); err != nil {
			logger.WithError(err).Error("r.body.Read() failed")
			return false
		}
	}
	return true
}

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

func jsonEncode(dest io.Writer, src interface{}) bool {
	if err := json.NewEncoder(dest).Encode(src); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		return false
	}
	return true
}

func jsonEncodeBeautify(dest io.Writer, src interface{}) bool {
	encoder := json.NewEncoder(dest)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(src); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		return false
	}
	return true
}

func gobEncode(dest io.Writer, src interface{}) bool {
	if err := gob.NewEncoder(dest).Encode(src); err != nil {
		logger.WithError(err).Error("gob.Encode() failed")
		return false
	}
	return true
}

func jsonDecode(src io.Reader, dest interface{}) bool {
	if err := json.NewDecoder(src).Decode(dest); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		return false
	}
	return true
}

func gobDecode(src io.Reader, dest interface{}) bool {
	if err := gob.NewDecoder(src).Decode(dest); err != nil {
		logger.WithError(err).Error("gob.Decode() failed")
		return false
	}
	return true
}

func acceptsType(r *http.Request, mediaType string) bool {
	for _, value := range r.Header["Accept"] {
		for _, wantType := range strings.Split(value, ",") {
			if strings.HasPrefix(mediaType, wantType) {
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

func notFound(w http.ResponseWriter) {
	http.Error(w, "404 page not found", http.StatusNotFound)
}

func unauthorized(w http.ResponseWriter) {
	http.Error(w, "authorization failed", http.StatusUnauthorized)
}

func notImplemented(w http.ResponseWriter) {
	http.Error(w, "", http.StatusNotImplemented)
}
