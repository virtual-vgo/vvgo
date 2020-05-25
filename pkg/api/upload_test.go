package api

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"io"
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
				PartNames:   []string{"trumpet 1"},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "sheet-music.pdf"),
				FileBytes:   sheetBytes,
				ContentType: "application/pdf",
			},
			want: ErrInvalidUploadType,
		},
		{
			name: "invalid part names",
			upload: Upload{
				UploadType:  UploadTypeSheets,
				PartNames:   []string{""},
				Project:     "01-snake-eater",
				FileName:    filepath.Join("testdata", "sheet-music.pdf"),
				FileBytes:   sheetBytes,
				ContentType: "application/pdf",
			},
			want: parts.ErrInvalidPartName,
		},
		{
			name: "invalid project",
			upload: Upload{
				UploadType:  UploadTypeClix,
				Project:     "00-mighty-morphin-power-rangers",
				PartNames:   []string{"trumpet 1"},
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
				PartNames:   []string{"trumpet 1"},
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
				PartNames:   []string{"trumpet 1"},
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
				PartNames:   []string{"trumpet 1"},
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
				PartNames:   []string{"trumpet 1"},
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
				PartNames:   []string{"trumpet 1"},
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
	readFile := func(name string) *bytes.Buffer {
		var dest bytes.Buffer
		file, err := os.Open(name)
		require.NoError(t, err, "os.Open() `%s`", file)
		_, err = dest.ReadFrom(file)
		require.NoError(t, err, "file.Read() `%s`", file)
		return &dest
	}

	// read test data from files
	sheetBytes := readFile("testdata/sheet-music.pdf")
	clickBytes := readFile("testdata/click-track.mp3")

	// create our own upload document
	upload := []Upload{
		{
			UploadType:  "sheets",
			PartNames:   []string{"trumpet 1"},
			Project:     "01-snake-eater",
			FileName:    "testdata/sheet-music.pdf",
			FileBytes:   sheetBytes.Bytes(),
			ContentType: "application/pdf",
		},
		{
			UploadType:  "clix",
			PartNames:   []string{"trumpet 1"},
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

	newRequest := func(t *testing.T, method string, url string, body io.Reader) *http.Request {
		req, err := http.NewRequest(method, url, body)
		assert.NoError(t, err, "http.NewRequest()")
		return req
	}

	t.Run("invalid method", func(t *testing.T) {
		ts := httptest.NewServer(UploadHandler{&Database{
			Parts:  newParts(),
			Distro: newBucket(t),
		}})
		defer ts.Close()

		req := newRequest(t, http.MethodGet, ts.URL, nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "http.Do()")
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("content-type:application/json", func(t *testing.T) {
		ts := httptest.NewServer(UploadHandler{&Database{
			Parts:  newParts(),
			Distro: newBucket(t),
		}})
		defer ts.Close()

		var uploadJSON bytes.Buffer
		require.NoError(t, json.NewEncoder(&uploadJSON).Encode(upload), "json.Encode()")

		req := newRequest(t, http.MethodPost, ts.URL, &uploadJSON)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "http.Do()")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var gotStatus []UploadStatus
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&gotStatus))
		assertEqualStatus(t, wantStatus, gotStatus)
	})

	t.Run("content-type:application/"+MediaTypeUploadsGob, func(t *testing.T) {
		ts := httptest.NewServer(UploadHandler{&Database{
			Parts:  newParts(),
			Distro: newBucket(t),
		}})
		defer ts.Close()

		var uploadGob bytes.Buffer
		require.NoError(t, gob.NewEncoder(&uploadGob).Encode(upload), "gob.Encode()")

		req := newRequest(t, http.MethodPost, ts.URL, &uploadGob)
		req.Header.Set("Content-Type", MediaTypeUploadsGob)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "http.Do()")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var gotStatus []UploadStatus
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&gotStatus))
		assertEqualStatus(t, wantStatus, gotStatus)
	})

	t.Run("content-encoding:gzip", func(t *testing.T) {
		ts := httptest.NewServer(UploadHandler{&Database{
			Parts:  newParts(),
			Distro: newBucket(t),
		}})
		defer ts.Close()

		var uploadGobGzip bytes.Buffer
		gzipWriter := gzip.NewWriter(&uploadGobGzip)
		require.NoError(t, gob.NewEncoder(gzipWriter).Encode(upload), "gob.Encode()")
		require.NoError(t, gzipWriter.Close(), "gzip.Close()")

		req := newRequest(t, http.MethodPost, ts.URL, &uploadGobGzip)
		req.Header.Set("Content-Type", MediaTypeUploadsGob)
		req.Header.Set("Content-Encoding", "application/gzip")
		req.Header.Set("Accept", MediaTypeUploadStatusesGob)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err, "http.Do()")
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var gotStatus []UploadStatus
		require.NoError(t, gob.NewDecoder(resp.Body).Decode(&gotStatus))
		assertEqualStatus(t, wantStatus, gotStatus)
	})
}

func assertEqualStatus(t *testing.T, want []UploadStatus, got []UploadStatus) {
	sort.Sort(statusSort(want))
	sort.Sort(statusSort(got))
	assert.Equal(t, want, got)
}

type statusSort []UploadStatus

func (x statusSort) Len() int           { return len(x) }
func (x statusSort) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
func (x statusSort) Less(i, j int) bool { return x[i].FileName < x[j].FileName }
