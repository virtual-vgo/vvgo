package errors

import (
	"encoding/json"
	"fmt"
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"net/http"
)

type Error struct {
	Code    int             `json:"Code"`
	Message string          `json:"Error"`
	Data    json.RawMessage `json:"Data"`
}

func (x Error) Error() string { return x.Message }

func NewJsonDecodeError(err error) http2.Response {
	return NewBadRequestError("invalid json: " + err.Error())
}

func NewBadRequestError(reason string) http2.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusBadRequest,
		Message: reason,
	})
}

func NewMethodNotAllowedError() http2.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusMethodNotAllowed,
		Message: "method not allowed",
	})
}

func NewUnauthorizedError() http2.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusUnauthorized,
		Message: "unauthorized",
	})
}

func NewNotFoundError(reason string) http2.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusNotFound,
		Message: reason,
	})
}

func NewInternalServerError() http2.Response {
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

func NewTooManyBytesError() http2.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusRequestEntityTooLarge,
		Message: "request too chonk",
	})
}

func NewNotImplementedError() http2.Response {
	return NewErrorResponse(Error{
		Code:    http.StatusNotImplemented,
		Message: "not implemented",
	})
}

func NewErrorResponse(error Error) http2.Response {
	return http2.Response{Status: http2.StatusError, Error: &error}
}
