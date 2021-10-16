package http_helpers

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"io"
	"net/http"
	"strings"
)

func JsonEncode(w http.ResponseWriter, src interface{}) bool {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(src); err != nil {
		logger.JsonEncodeFailure(context.Background(), err)
		return false
	}
	return true
}

func JsonDecodeFailure(ctx context.Context, w http.ResponseWriter, err error) {
	logger.JsonDecodeFailure(ctx, err)
	BadRequest(ctx, w, "invalid json: "+err.Error())
}

func BadRequest(ctx context.Context, w http.ResponseWriter, reason string) {
	WriteErrorResponse(ctx, w, models.ErrorResponse{
		Code:  http.StatusBadRequest,
		Error: reason,
	})
}

func MethodNotAllowed(ctx context.Context, w http.ResponseWriter) {
	WriteErrorResponse(ctx, w, models.ErrorResponse{
		Code:  http.StatusMethodNotAllowed,
		Error: "method not allowed",
	})
}

func Unauthorized(ctx context.Context, w http.ResponseWriter) {
	WriteErrorResponse(ctx, w, models.ErrorResponse{
		Code:  http.StatusUnauthorized,
		Error: "unauthorized",
	})
}

func InternalServerError(ctx context.Context, w http.ResponseWriter) {
	WriteErrorResponse(ctx, w, models.ErrorResponse{
		Code:  http.StatusInternalServerError,
		Error: "internal server error",
	})
}

func WriteErrorResponse(ctx context.Context, w http.ResponseWriter, error models.ErrorResponse) {
	WriteAPIResponse(ctx, w, models.Response{
		Status: models.StatusError,
		Type:   models.ResponseTypeError,
		Error:  &error,
	})
}

func WriteAPIResponse(_ context.Context, w http.ResponseWriter, resp models.Response) {
	var code int
	switch {
	case resp.Status != models.StatusError:
		code = 200
	case resp.Error != nil:
		code = resp.Error.Code
	default:
		code = http.StatusInternalServerError
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&resp)
}

func JsonDecode(src io.Reader, dest interface{}) bool {
	if err := json.NewDecoder(src).Decode(dest); err != nil {
		logger.JsonDecodeFailure(context.Background(), err)
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

func TooManyBytes(w http.ResponseWriter)   { http.Error(w, "", http.StatusRequestEntityTooLarge) }
func InvalidContent(w http.ResponseWriter) { http.Error(w, "", http.StatusUnsupportedMediaType) }
func NotFound(w http.ResponseWriter)       { http.Error(w, "404 page not found", http.StatusNotFound) }
func NotImplemented(w http.ResponseWriter) { http.Error(w, "", http.StatusNotImplemented) }
