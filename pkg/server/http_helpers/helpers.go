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

func NewOkResponse() models.ApiResponse {
	return models.ApiResponse{Status: models.StatusOk}
}

func NewJsonDecodeError(err error) models.ApiResponse {
	return NewBadRequestError("invalid json: " + err.Error())
}

func NewBadRequestError(reason string) models.ApiResponse {
	return NewErrorResponse(models.ApiError{
		Code:  http.StatusBadRequest,
		Error: reason,
	})
}

func NewMethodNotAllowedError() models.ApiResponse {
	return NewErrorResponse(models.ApiError{
		Code:  http.StatusMethodNotAllowed,
		Error: "method not allowed",
	})
}

func NewUnauthorizedError() models.ApiResponse {
	return NewErrorResponse(models.ApiError{
		Code:  http.StatusUnauthorized,
		Error: "unauthorized",
	})
}

func NewNotFoundError(reason string) models.ApiResponse {
	return NewErrorResponse(models.ApiError{
		Code:  http.StatusNotFound,
		Error: reason,
	})
}

func NewInternalServerError() models.ApiResponse {
	return NewErrorResponse(models.ApiError{
		Code:  http.StatusInternalServerError,
		Error: "internal server error",
	})
}

func WriteErrorJsonDecodeFailure(ctx context.Context, w http.ResponseWriter, err error) {
	logger.JsonDecodeFailure(ctx, err)
	WriteAPIResponse(ctx, w, NewJsonDecodeError(err))
}

func WriteErrorBadRequest(ctx context.Context, w http.ResponseWriter, reason string) {
	WriteAPIResponse(ctx, w, NewBadRequestError(reason))
}

func NewErrorResponse(error models.ApiError) models.ApiResponse {
	return models.ApiResponse{Status: models.StatusError, Error: &error}
}

func WriteErrorMethodNotAllowed(ctx context.Context, w http.ResponseWriter) {
	WriteAPIResponse(ctx, w, NewMethodNotAllowedError())
}

func WriteUnauthorizedError(ctx context.Context, w http.ResponseWriter) {
	WriteAPIResponse(ctx, w, NewUnauthorizedError())
}
func WriteInternalServerError(ctx context.Context, w http.ResponseWriter) {
	WriteAPIResponse(ctx, w, NewInternalServerError())
}

func WriteErrorResponse(ctx context.Context, w http.ResponseWriter, error models.ApiError) {
	WriteAPIResponse(ctx, w, NewErrorResponse(error))
}

func WriteErrorTooManyBytes(ctx context.Context, w http.ResponseWriter) {
	WriteErrorResponse(ctx, w, models.ApiError{
		Code:  http.StatusRequestEntityTooLarge,
		Error: "request too chonk",
	})
}

func WriteErrorUnsupportedFile(ctx context.Context, w http.ResponseWriter) {
	WriteErrorResponse(ctx, w, models.ApiError{
		Code:  http.StatusUnsupportedMediaType,
		Error: "unsupported file",
	})
}

func WriteErrorNotFound(ctx context.Context, w http.ResponseWriter) {
	WriteErrorResponse(ctx, w, models.ApiError{
		Code:  http.StatusNotFound,
		Error: "not found",
	})
}

func WriteErrorNotImplemented(ctx context.Context, w http.ResponseWriter) {
	WriteErrorResponse(ctx, w, models.ApiError{
		Code:  http.StatusNotImplemented,
		Error: "not implemented",
	})
}

func WriteAPIResponse(_ context.Context, w http.ResponseWriter, resp models.ApiResponse) {
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
