package api

import (
	"bytes"
	"compress/gzip"
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
	"sort"
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
				PartNames:   []string{""},
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
	warehouse, err := storage.NewWarehouse(storage.Config{})
	require.NoError(t, err, "storage.NewWarehouse()")

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
	var uploadGob, uploadGobGzip, uploadJSON, wantStatusGob, wantStatusJSON bytes.Buffer
	require.NoError(t, gob.NewEncoder(&uploadGob).Encode(upload), "gob.Encode()")
	require.NoError(t, json.NewEncoder(&uploadJSON).Encode(upload), "json.Encode()")
	require.NoError(t, gob.NewEncoder(&wantStatusGob).Encode(wantStatus), "gob.Encode()")
	require.NoError(t, json.NewEncoder(&wantStatusJSON).Encode(wantStatus), "json.Encode()")
	gzipWriter := gzip.NewWriter(&uploadGobGzip)
	_, err = gzipWriter.Write(uploadGob.Bytes())
	require.NoError(t, err, "gzip.Write()")
	require.NoError(t, gzipWriter.Close(), "gzip.Close()")

	type request struct {
		method    string
		mediaType string
		accept    string
		encoding  string
		body      bytes.Buffer
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
				method:    http.MethodGet,
				mediaType: "application/json",
				body:      *bytes.NewBuffer([]byte("[]")),
			},
			wants: wants{
				body: *bytes.NewBuffer([]byte("\n")),
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "content-type:application/json/success",
			request: request{
				method:    http.MethodPost,
				mediaType: "application/json",
				accept:    "application/json",
				body:      uploadJSON,
			},
			wants: wants{
				body: wantStatusJSON,
				code: http.StatusOK,
			},
		},
		{
			name: "content-type:application/" + MediaTypeUploadsGob + "/success",
			request: request{
				method:    http.MethodPost,
				mediaType: MediaTypeUploadsGob,
				accept:    MediaTypeUploadStatusesGob,
				body:      uploadGob,
			},
			wants: wants{
				body: wantStatusGob,
				code: http.StatusOK,
			},
		},
		{
			name: "content-encoding/gzip/success",
			request: request{
				method:    http.MethodPost,
				mediaType: MediaTypeUploadsGob,
				accept:    MediaTypeUploadStatusesGob,
				encoding:  "application/gzip",
				body:      uploadGobGzip,
			},
			wants: wants{
				body: wantStatusGob,
				code: http.StatusOK,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {

			ctx := context.Background()
			bucket, err := warehouse.NewBucket(ctx, "testing")
			require.NoError(t, err, "storage.NewBucket")
			handlerStorage := Storage{
				Parts:  newParts(),
				Sheets: bucket,
				Clix:   bucket,
				Tracks: bucket,
				StorageConfig: StorageConfig{
					SheetsBucketName: "sheets",
					ClixBucketName:   "clix",
					TracksBucketName: "tracks",
				},
			}

			request := httptest.NewRequest(tt.request.method, "/upload", &tt.request.body)
			request.Header.Set("Content-Type", tt.request.mediaType)
			request.Header.Set("Content-Encoding", tt.request.encoding)
			request.Header.Set("Accept", tt.request.accept)
			recorder := httptest.NewRecorder()
			UploadHandler{&handlerStorage}.ServeHTTP(recorder, request)
			resp := recorder.Result()
			var respBody bytes.Buffer
			respBody.ReadFrom(resp.Body)
			resp.Body.Close()
			assert.Equal(t, tt.wants.code, resp.StatusCode, "code")
			if !assert.Equal(t, tt.wants.body.String(), respBody.String(), "body") {
				var gotStatus []UploadStatus
				gob.NewDecoder(recorder.Body).Decode(&gotStatus)
				sort.Sort(statusSort(wantStatus))
				sort.Sort(statusSort(gotStatus))
				assert.Equal(t, wantStatus, gotStatus)
			}
		})
	}
}

type statusSort []UploadStatus

func (x statusSort) Len() int           { return len(x) }
func (x statusSort) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x statusSort) Less(i, j int) bool { return x[i].FileName < x[j].FileName }
