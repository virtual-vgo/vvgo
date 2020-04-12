package api

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadHandler_ServeHTTP(t *testing.T) {
	type wants struct {
		code int
		body string
	}

	sheetsUploadBody, err := ioutil.ReadFile("testdata/sheet-upload.json")
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}

	clixUploadBody, err := ioutil.ReadFile("testdata/clix-upload.json")
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}

	type request struct {
		method      string
		contentType string
		body        string
	}

	mocks := struct {
		bucket MockBucket
		locker MockLocker
	}{
		bucket: MockBucket{
			getObject: func(_ string, object *storage.Object) bool {
				*object = storage.Object{ContentType: "", Buffer: *bytes.NewBuffer([]byte(`[]`))}
				return true
			},
			putObject: func(string, *storage.Object) bool { return true },
			putFile: func(file *storage.File) bool {
				return true
			},
		},
		locker: MockLocker{
			lock:   func(ctx context.Context) bool { return true },
			unlock: func() {},
		},
	}

	for _, tt := range []struct {
		name    string
		request request
		wants   wants
	}{
		{
			name: "method:get",
			request: request{
				method:      http.MethodGet,
				contentType: "application/json",
				body:        string(sheetsUploadBody),
			},
			wants: wants{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "content-type:text/html",
			request: request{
				method:      http.MethodPost,
				contentType: "text/html",
				body:        string(sheetsUploadBody),
			},
			wants: wants{
				code: http.StatusUnsupportedMediaType,
			},
		},
		{
			name: "body:invalid-json",
			request: request{
				method:      http.MethodPost,
				contentType: "application/json",
				body:        `invalid-json`,
			},
			wants: wants{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "type:sheets/success",
			request: request{
				method:      http.MethodPost,
				contentType: "application/json",
				body:        string(sheetsUploadBody),
			},
			wants: wants{
				body: `[{"file_name":"music-sheet.pdf","code":200}]`,
				code: http.StatusOK,
			},
		},
		{
			name: "type:clix/success",
			request: request{
				method:      http.MethodPost,
				contentType: "application/json",
				body:        string(clixUploadBody),
			},
			wants: wants{
				body: `[{"file_name":"click-track.mp3","code":200}]`,
				code: http.StatusOK,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.method, "/upload", strings.NewReader(tt.request.body))
			request.Header.Set("Content-Type", tt.request.contentType)
			recorder := httptest.NewRecorder()
			UploadHandler{&Database{
				Parts: parts.Parts{
					Bucket: &mocks.bucket,
					Locker: &mocks.locker,
				},
				Sheets: &mocks.bucket,
				Clix:   &mocks.bucket,
			}}.ServeHTTP(recorder, request)
			assert.Equal(t, tt.wants.code, recorder.Code, "code")
			assert.Equal(t, tt.wants.body, strings.TrimSpace(recorder.Body.String()), "body")
		})
	}
}

func TestUpload_ValidateSheets(t *testing.T) {

}

func TestUpload_Sheets(t *testing.T) {

}
