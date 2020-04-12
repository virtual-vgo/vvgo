package api

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpload_Validate(t *testing.T) {
	sheetBytes, err := ioutil.ReadFile(filepath.Join("testdata", "sheet-music.pdf"))
	require.NoError(t, err, "ioutil.ReadFile() failed")
	clickBytes, err := ioutil.ReadFile(filepath.Join("testdata", "click-track.mp3"))
	require.NoError(t, err, "ioutil.ReadFile() failed")

	for _, tt := range []struct {
		name   string
		upload Upload
		want   error
	}{
		{
			name: "invalid type",
			upload: Upload{
				PartNames:   []string{"trumpet"},
				PartNumbers: []uint8{1},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "sheet-music.pdf"),
				FileBytes:   sheetBytes,
				ContentType: "application/pdf",
			},
			want: ErrInvalidUploadType,
		},
		{
			name: "missing part names",
			upload: Upload{
				UploadType:  UploadTypeSheets,
				PartNumbers: []uint8{1},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "sheet-music.pdf"),
				FileBytes:   sheetBytes,
				ContentType: "application/pdf",
			},
			want: ErrMissingPartNames,
		},
		{
			name: "invalid part names",
			upload: Upload{
				UploadType:  UploadTypeSheets,
				PartNames:   []string{"not-an-instrument"},
				PartNumbers: []uint8{1},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "sheet-music.pdf"),
				FileBytes:   sheetBytes,
				ContentType: "application/pdf",
			},
			want: parts.ErrInvalidPartName,
		},
		{
			name: "missing part numbers",
			upload: Upload{
				UploadType:  UploadTypeClix,
				PartNames:   []string{"trumpet"},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "click-track.mp3"),
				FileBytes:   clickBytes,
				ContentType: "audio/mpeg",
			},
			want: ErrMissingPartNumbers,
		},
		{
			name: "invalid part numbers",
			upload: Upload{
				UploadType:  UploadTypeClix,
				PartNames:   []string{"trumpet"},
				PartNumbers: []uint8{0},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "click-track.mp3"),
				FileBytes:   clickBytes,
				ContentType: "audio/mpeg",
			},
			want: parts.ErrInvalidPartNumber,
		},
		{
			name: "invalid project",
			upload: Upload{
				UploadType:  UploadTypeClix,
				Project:     "00-mighty-morphin-power-ranger",
				PartNames:   []string{"trumpet"},
				PartNumbers: []uint8{1},
				FileName:    filepath.Join("testdata", "click-track.mp3"),
				FileBytes:   clickBytes,
				ContentType: "audio/mpeg",
			},
			want: projects.ErrNotFound,
		},
		{
			name: "empty file bytes",
			upload: Upload{
				UploadType:  UploadTypeClix,
				PartNames:   []string{"trumpet"},
				PartNumbers: []uint8{1},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "click-track.mp3"),
				ContentType: "audio/mpeg",
			},
			want: ErrEmptyFileBytes,
		},
		{
			name: "click/invalid",
			upload: Upload{
				UploadType:  UploadTypeClix,
				PartNames:   []string{"trumpet"},
				PartNumbers: []uint8{1},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "click-track.mp3"),
				FileBytes:   sheetBytes,
				ContentType: "audio/mpeg",
			},
			want: storage.ErrDetectedInvalidContent,
		},
		{
			name: "click/valid",
			upload: Upload{
				UploadType:  UploadTypeClix,
				PartNames:   []string{"trumpet"},
				PartNumbers: []uint8{1},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "click-track.mp3"),
				FileBytes:   clickBytes,
				ContentType: "audio/mpeg",
			},
			want: nil,
		},
		{
			name: "sheet/invalid",
			upload: Upload{
				UploadType:  UploadTypeSheets,
				PartNames:   []string{"trumpet"},
				PartNumbers: []uint8{1},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "sheet-music.pdf"),
				FileBytes:   clickBytes,
				ContentType: "application/pdf",
			},
			want: storage.ErrDetectedInvalidContent,
		},
		{
			name: "sheet/valid",
			upload: Upload{
				UploadType:  UploadTypeSheets,
				PartNames:   []string{"trumpet"},
				PartNumbers: []uint8{1},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "sheet-music.pdf"),
				FileBytes:   sheetBytes,
				ContentType: "application/pdf",
			},
			want: nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.upload.Validate())
		})
	}
}

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
			name: "type:sheets/failure",
			request: request{
				method:      http.MethodPost,
				contentType: "application/json",
				body:        "garbage",
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
