package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_acceptsType(t *testing.T) {
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
			if expected, got := tt.want, acceptsType(&http.Request{Header: tt.header}, tt.arg); expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}

func Test_badRequest(t *testing.T) {
	recorder := httptest.NewRecorder()
	badRequest(recorder, "some-reason")
	wantBody := "some-reason"
	wantCode := http.StatusBadRequest

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func Test_internalServerError(t *testing.T) {
	recorder := httptest.NewRecorder()
	internalServerError(recorder)
	wantBody := ""
	wantCode := http.StatusInternalServerError

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func Test_invalidContent(t *testing.T) {
	recorder := httptest.NewRecorder()
	invalidContent(recorder)
	wantBody := ""
	wantCode := http.StatusUnsupportedMediaType

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func Test_methodNotAllowed(t *testing.T) {
	recorder := httptest.NewRecorder()
	methodNotAllowed(recorder)
	wantBody := ""
	wantCode := http.StatusMethodNotAllowed

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func Test_notFound(t *testing.T) {
	recorder := httptest.NewRecorder()
	notFound(recorder)
	wantBody := "404 page not found"
	wantCode := http.StatusNotFound

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func Test_tooManyBytes(t *testing.T) {
	recorder := httptest.NewRecorder()
	tooManyBytes(recorder)
	wantBody := ""
	wantCode := http.StatusRequestEntityTooLarge

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func Test_unauthorized(t *testing.T) {
	recorder := httptest.NewRecorder()
	unauthorized(recorder)
	wantBody := "authorization failed"
	wantCode := http.StatusUnauthorized

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}

func Test_notImplemented(t *testing.T) {
	recorder := httptest.NewRecorder()
	notImplemented(recorder)
	wantBody := ""
	wantCode := http.StatusNotImplemented

	assert.Equal(t, wantCode, recorder.Code, "response code")
	assert.Equal(t, wantBody, strings.TrimSpace(recorder.Body.String()), "response body")
}
