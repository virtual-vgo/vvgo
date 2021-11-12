package response

import (
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"net/http"
)

type Error struct {
	Code    int             `json:"Code"`
	Message string          `json:"Error"`
	Data    json.RawMessage `json:"Data"`
}

func (x Error) Error() string { return x.Message }

func NewJsonDecodeError(err error) api.Response {
	return NewBadRequestError("invalid json: " + err.Error())
}

func NewBadRequestError(reason string) api.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusBadRequest,
		Message: reason,
	})
}

func NewMethodNotAllowedError() api.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusMethodNotAllowed,
		Message: "method not allowed",
	})
}

func NewUnauthorizedError() api.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusUnauthorized,
		Message: "unauthorized",
	})
}

func NewNotFoundError(reason string) api.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusNotFound,
		Message: reason,
	})
}

func NewInternalServerError() api.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusInternalServerError,
		Message: "internal server error",
	})
}

var InternalServerError = Error{
	Code:    http.StatusInternalServerError,
	Message: "internal server error",
}

func RedisError(err error) Error {
	return Error{
		Code:    http.StatusInternalServerError,
		Message: fmt.Sprintf("redis error: %s", err),
	}
}

func NewTooManyBytesError() api.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusRequestEntityTooLarge,
		Message: "request too chonk",
	})
}

func NewNotImplementedError() api.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusNotImplemented,
		Message: "not implemented",
	})
}

func NewErrorResponse(error Error) api.Response {
	return api.Response{Status: api.StatusError, Error: &error}
}
