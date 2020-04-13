package api

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

	// all the mocks always return true
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

	// read test data from files
	var sheetBytes, clickBytes bytes.Buffer
	for _, args := range []struct {
		buffer *bytes.Buffer
		file   string
	}{
		{&sheetBytes, "testdata/sheet-music.pdf"},
		{&clickBytes, "testdata/click-track.mp3"},
	} {
		file, err := os.Open(args.file)
		require.NoError(t, err, "os.Open()  `%s`", file)
		_, err = args.buffer.ReadFrom(file)
		require.NoError(t, err, "file.Read() `%s`", file)
	}

	// create our own upload document
	upload := []Upload{
		{
			UploadType:  "sheets",
			PartNames:   []string{"trumpet"},
			PartNumbers: []uint8{1},
			Project:     "01-snake-eater",
			FileName:    "testdata/sheet-music.pdf",
			FileBytes:   sheetBytes.Bytes(),
			ContentType: "application/pdf",
		},
		{
			UploadType:  "clix",
			PartNames:   []string{"trumpet"},
			PartNumbers: []uint8{1},
			Project:     "01-snake-eater",
			FileName:    "testdata/click-track.mp3",
			FileBytes:   clickBytes.Bytes(),
			ContentType: "audio/mpeg",
		},
	}

	wantStatus := []UploadStatus{
		{
			FileName: "testdata/sheet-music.pdf",
			Code:     http.StatusOK,
		},
		{
			FileName: "testdata/click-track.mp3",
			Code:     http.StatusOK,
		},
	}

	// encode it in our various encodings
	var uploadGob, uploadJSON, wantStatusGob, wantStatusJSON bytes.Buffer
	require.NoError(t, gob.NewEncoder(&uploadGob).Encode(upload), "gob.Encode()")
	require.NoError(t, json.NewEncoder(&uploadJSON).Encode(upload), "json.Encode()")
	require.NoError(t, gob.NewEncoder(&wantStatusGob).Encode(wantStatus), "gob.Encode()")
	require.NoError(t, json.NewEncoder(&wantStatusJSON).Encode(wantStatus), "json.Encode()")

	type request struct {
		method      string
		contentType string
		body        bytes.Buffer
	}

	type wants struct {
		code int
		body bytes.Buffer
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
				body:        *bytes.NewBuffer([]byte("[]")),
			},
			wants: wants{
				body: *bytes.NewBuffer([]byte("\n")),
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "content-type:text/html/failure",
			request: request{
				method:      http.MethodPost,
				contentType: "text/html",
				body:        *bytes.NewBuffer([]byte("")),
			},
			wants: wants{
				body: *bytes.NewBuffer([]byte("\n")),
				code: http.StatusUnsupportedMediaType,
			},
		},
		{
			name: "content-type:application/json/success",
			request: request{
				method:      http.MethodPost,
				contentType: "application/json",
				body:        uploadJSON,
			},
			wants: wants{
				body: wantStatusJSON,
				code: http.StatusOK,
			},
		},
		{
			name: "content-type:application/octet-stream/success",
			request: request{
				method:      http.MethodPost,
				contentType: "application/json",
				body:        uploadGob,
			},
			wants: wants{
				body: wantStatusGob,
				code: http.StatusOK,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.method, "/upload", &tt.request.body)
			request.Header.Set("Content-Type", tt.request.contentType)
			recorder := httptest.NewRecorder()
			UploadHandler{&Storage{
				Parts: parts.Parts{
					Bucket: &mocks.bucket,
					Locker: &mocks.locker,
				},
				Sheets: &mocks.bucket,
				Clix:   &mocks.bucket,
			}}.ServeHTTP(recorder, request)
			assert.Equal(t, tt.wants.code, recorder.Code, "code")
			assert.Equal(t, tt.wants.body.String(), recorder.Body.String(), "body")
		})
	}
}
