package errors

import (
	"context"
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/test_helpers"
	"net/http"
	"net/http/httptest"
	"testing"
)

var ctx = context.Background()

func TestBadRequest(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewBadRequestError("some-reason").WriteHTTP(ctx, recorder, httptest.NewRequest("GET", "/", nil))
	test_helpers.AssertEqualResponse(t, http2.Response{
		Status: http2.StatusError,
		Error: &Error{
			Code:    http.StatusBadRequest,
			Message: "some-reason",
		},
	}, recorder.Result())
}

func TestInternalServerError(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewInternalServerError().WriteHTTP(ctx, recorder, httptest.NewRequest("GET", "/", nil))
	test_helpers.AssertEqualResponse(t, http2.Response{
		Status: http2.StatusError,
		Error: &Error{
			Code:    http.StatusInternalServerError,
			Message: "internal server error",
		},
	}, recorder.Result())
}

func TestMethodNotAllowed(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewMethodNotAllowedError().WriteHTTP(ctx, recorder, httptest.NewRequest("GET", "/", nil))
	test_helpers.AssertEqualResponse(t, http2.Response{
		Status: http2.StatusError,
		Error: &Error{
			Code:    http.StatusMethodNotAllowed,
			Message: "method not allowed",
		},
	}, recorder.Result())
}

func TestNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewNotFoundError("not found").WriteHTTP(ctx, recorder, httptest.NewRequest("GET", "/", nil))
	test_helpers.AssertEqualResponse(t, http2.Response{
		Status: http2.StatusError,
		Error: &Error{
			Code:    http.StatusNotFound,
			Message: "not found",
		},
	}, recorder.Result())
}

func TestTooManyBytes(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewTooManyBytesError().WriteHTTP(ctx, recorder, httptest.NewRequest("GET", "/", nil))
	test_helpers.AssertEqualResponse(t, http2.Response{
		Status: http2.StatusError,
		Error: &Error{
			Code:    http.StatusRequestEntityTooLarge,
			Message: "request too chonk",
		},
	}, recorder.Result())
}

func TestUnauthorized(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewUnauthorizedError().WriteHTTP(ctx, recorder, httptest.NewRequest("GET", "/", nil))
	test_helpers.AssertEqualResponse(t, http2.Response{
		Status: http2.StatusError,
		Error: &Error{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized",
		},
	}, recorder.Result())
}

func TestNotImplemented(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewNotImplementedError().WriteHTTP(ctx, recorder, httptest.NewRequest("GET", "/", nil))
	test_helpers.AssertEqualResponse(t, http2.Response{
		Status: http2.StatusError,
		Error: &Error{
			Code:    http.StatusNotImplemented,
			Message: "not implemented",
		},
	}, recorder.Result())
}
