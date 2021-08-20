package helpers

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	BadRequest(recorder, "some-reason")
	wantBody := "some-reason"
	wantCode := http.StatusBadRequest

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func TestInternalServerError(t *testing.T) {
	recorder := httptest.NewRecorder()
	InternalServerError(recorder)
	wantBody := ""
	wantCode := http.StatusInternalServerError

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func TestInvalidContent(t *testing.T) {
	recorder := httptest.NewRecorder()
	InvalidContent(recorder)
	wantBody := ""
	wantCode := http.StatusUnsupportedMediaType

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func TestMethodNotAllowed(t *testing.T) {
	recorder := httptest.NewRecorder()
	MethodNotAllowed(recorder)
	wantBody := ""
	wantCode := http.StatusMethodNotAllowed

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func TestNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	NotFound(recorder)
	wantBody := "404 page not found"
	wantCode := http.StatusNotFound

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func TestTooManyBytes(t *testing.T) {
	recorder := httptest.NewRecorder()
	TooManyBytes(recorder)
	wantBody := ""
	wantCode := http.StatusRequestEntityTooLarge

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func TestUnauthorized(t *testing.T) {
	recorder := httptest.NewRecorder()
	Unauthorized(recorder)
	wantBody := "authorization failed"
	wantCode := http.StatusUnauthorized

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func TestNotImplemented(t *testing.T) {
	recorder := httptest.NewRecorder()
	NotImplemented(recorder)
	wantBody := ""
	wantCode := http.StatusNotImplemented

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}
