package helpers

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"io"
	"net/http"
	"strings"
)

func JsonEncode(w http.ResponseWriter, src interface{}) bool {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(src); err != nil {
		logger.WithError(err).Error("json.Encode() failed")
		return false
	}
	return true
}

func JsonDecode(src io.Reader, dest interface{}) bool {
	if err := json.NewDecoder(src).Decode(dest); err != nil {
		logger.WithError(err).Error("json.Decode() failed")
		return false
	}
	return true
}

func AcceptsType(r *http.Request, mediaType string) bool {
	for _, value := range r.Header["Accept"] {
		for _, wantType := range strings.Split(value, ",") {
			if strings.HasPrefix(mediaType, wantType) {
				return true
			}
		}
	}
	return false
}

func BadRequest(w http.ResponseWriter, reason string) { http.Error(w, reason, http.StatusBadRequest) }
func InternalServerError(w http.ResponseWriter)       { http.Error(w, "", http.StatusInternalServerError) }
func MethodNotAllowed(w http.ResponseWriter)          { http.Error(w, "", http.StatusMethodNotAllowed) }
func TooManyBytes(w http.ResponseWriter)              { http.Error(w, "", http.StatusRequestEntityTooLarge) }
func InvalidContent(w http.ResponseWriter)            { http.Error(w, "", http.StatusUnsupportedMediaType) }
func NotFound(w http.ResponseWriter)                  { http.Error(w, "404 page not found", http.StatusNotFound) }
func Unauthorized(w http.ResponseWriter) {
	http.Error(w, "authorization failed", http.StatusUnauthorized)
}
func NotImplemented(w http.ResponseWriter) { http.Error(w, "", http.StatusNotImplemented) }
