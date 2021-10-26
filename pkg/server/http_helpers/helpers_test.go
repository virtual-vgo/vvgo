package http_helpers

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers/test_helpers"
	"net/http"
	"net/http/httptest"
	"testing"
)

var ctx = context.Background()

func TestAcceptsType(t *testing.T) {
	for _, tt := range []struct {
		name   string
		header http.Header
		arg    string
		want   bool
	}{
		{
			name: "yep",
			header: http.Header{
				"Accept": []string{"text/html,application/json,application/pdf", "application/xml,cheese/sandwich"},
			},
			arg:  "application/xml",
			want: true,
		},
		{
			name: "nope",
			header: http.Header{
				"Accept": []string{"text/html,application/json,application/pdf", "application/xml,cheese/sandwich"},
			},
			arg: "sour/cream", want: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if expected, got := tt.want, AcceptsType(&http.Request{Header: tt.header}, tt.arg); expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}

func TestBadRequest(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteErrorBadRequest(ctx, recorder, "some-reason")
	test_helpers.AssertEqualResponse(t, models.ApiResponse{
		Status: models.StatusError,
		Error: &models.ApiError{
			Code:  http.StatusBadRequest,
			Error: "some-reason",
		},
	}, recorder.Result())
}

func TestInternalServerError(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteInternalServerError(ctx, recorder)
	test_helpers.AssertEqualResponse(t, models.ApiResponse{
		Status: models.StatusError,
		Error: &models.ApiError{
			Code:  http.StatusInternalServerError,
			Error: "internal server error",
		},
	}, recorder.Result())
}

func TestUnsupportedFile(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteErrorUnsupportedFile(ctx, recorder)
	test_helpers.AssertEqualResponse(t, models.ApiResponse{
		Status: models.StatusError,
		Error: &models.ApiError{
			Code:  http.StatusUnsupportedMediaType,
			Error: "unsupported file",
		},
	}, recorder.Result())
}

func TestMethodNotAllowed(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteErrorMethodNotAllowed(ctx, recorder)
	test_helpers.AssertEqualResponse(t, models.ApiResponse{
		Status: models.StatusError,
		Error: &models.ApiError{
			Code:  http.StatusMethodNotAllowed,
			Error: "method not allowed",
		},
	}, recorder.Result())
}

func TestNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteErrorNotFound(ctx, recorder)
	test_helpers.AssertEqualResponse(t, models.ApiResponse{
		Status: models.StatusError,
		Error: &models.ApiError{
			Code:  http.StatusNotFound,
			Error: "not found",
		},
	}, recorder.Result())
}

func TestTooManyBytes(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteErrorTooManyBytes(ctx, recorder)
	test_helpers.AssertEqualResponse(t, models.ApiResponse{
		Status: models.StatusError,
		Error: &models.ApiError{
			Code:  http.StatusRequestEntityTooLarge,
			Error: "request too chonk",
		},
	}, recorder.Result())
}

func TestUnauthorized(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteUnauthorizedError(ctx, recorder)
	test_helpers.AssertEqualResponse(t, models.ApiResponse{
		Status: models.StatusError,
		Error: &models.ApiError{
			Code:  http.StatusUnauthorized,
			Error: "unauthorized",
		},
	}, recorder.Result())
}

func TestNotImplemented(t *testing.T) {
	recorder := httptest.NewRecorder()
	WriteErrorNotImplemented(ctx, recorder)
	test_helpers.AssertEqualResponse(t, models.ApiResponse{
		Status: models.StatusError,
		Error: &models.ApiError{
			Code:  http.StatusNotImplemented,
			Error: "not implemented",
		},
	}, recorder.Result())
}
