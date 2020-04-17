package api

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPartsHandler_ServeHTTP(t *testing.T) {
	clixBucket := "clix"
	sheetsBucket := "sheets"

	type request struct {
		method  string
		body    string
		accepts string
	}
	type wants struct {
		code int
		body string
	}
	fileBytes, err := ioutil.ReadFile("testdata/parts.html")
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	mockHTML := string(fileBytes)
	fileBytes, err = ioutil.ReadFile("testdata/parts.json")
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	mockJSON := string(fileBytes)
	mockBucket := MockBucket{getObject: func(name string, dest *storage.Object) bool {
		if name == parts.DataFile {
			parts := []parts.Part{
				{
					ID: parts.ID{
						Project: "01-snake-eater",
						Name:    "trumpet",
						Number:  3,
					},
					Sheets: []parts.Link{{ObjectKey: "sheet.pdf", CreatedAt: time.Now()}},
					Clix:   []parts.Link{{ObjectKey: "click.mp3", CreatedAt: time.Now()}},
				},
				{
					ID: parts.ID{
						Project: "02-proof-of-a-hero",
						Name:    "trumpet",
						Number:  3,
					},
					Sheets: []parts.Link{{ObjectKey: "sheet.pdf", CreatedAt: time.Now()}},
					Clix:   []parts.Link{{ObjectKey: "click.mp3", CreatedAt: time.Now()}},
				},
			}
			var buffer bytes.Buffer
			json.NewEncoder(&buffer).Encode(parts)
			*dest = storage.Object{
				ContentType: "application/json",
				Buffer:      buffer,
			}
		}
		return true
	}}

	for _, tt := range []struct {
		name    string
		request request
		wants   wants
	}{
		{
			name: "method post",
			request: request{
				method: http.MethodPost,
			},
			wants: wants{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "method get",
			request: request{
				method: http.MethodGet,
			},
			wants: wants{
				code: http.StatusOK,
				body: mockJSON,
			},
		},
		{
			name: "method get/accept text/html",
			request: request{
				method:  http.MethodGet,
				accepts: "text/html",
				body:    mockHTML,
			},
			wants: wants{
				code: http.StatusOK,
				body: mockHTML,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := PartsHandler{NavBar{}, &Storage{
				Parts:  parts.Parts{Bucket: &mockBucket},
				Sheets: &mockBucket,
				Clix:   &mockBucket,
				ServerConfig: ServerConfig{
					SheetsBucketName: sheetsBucket,
					ClixBucketName:   clixBucket,
				},
			}}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.request.method, "/sheets", strings.NewReader(tt.request.body))
			request.Header.Set("Accept", tt.request.accepts)
			server.ServeHTTP(recorder, request)
			gotResp := recorder.Result()
			gotBody := strings.TrimSpace(recorder.Body.String())
			t.Logf("Expected body:\n%s\n", tt.wants.body)
			t.Logf("Got body:\n%s\n", gotBody)
			if expected, got := tt.wants.code, gotResp.StatusCode; expected != got {
				t.Errorf("expected code %v, got %v", expected, got)
			}

			switch tt.wants.body {
			case mockJSON:
				var buf []map[string]interface{}
				json.Unmarshal([]byte(mockJSON), &buf)
				wantBytes, _ := json.Marshal(&buf)
				tt.wants.body = string(wantBytes)
				buf = nil
				json.Unmarshal([]byte(gotBody), &buf)
				gotBytes, err := json.Marshal(&buf)
				require.NoError(t, err, "json.Unmarshal")
				gotBody = string(gotBytes)

			case mockHTML:
				m := minify.New()
				m.AddFunc("text/html", html.Minify)
				var gotBuf bytes.Buffer
				if err := m.Minify("text/html", &gotBuf, strings.NewReader(gotBody)); err != nil {
					panic(err)
				}
				gotBody = gotBuf.String()

				var wantBuf bytes.Buffer
				if err := m.Minify("text/html", &wantBuf, strings.NewReader(tt.wants.body)); err != nil {
					panic(err)
				}
				tt.wants.body = wantBuf.String()
			}

			assert.Equal(t, tt.wants.body, gotBody, "body")
		})
	}
}

func TestIndexHandler_ServeHTTP(t *testing.T) {
	wantCode := http.StatusOK
	wantBytes, err := ioutil.ReadFile("testdata/index.html")
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}

	server := IndexHandler{}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	server.ServeHTTP(recorder, request)
	gotResp := recorder.Result()
	t.Logf("Got Body:\n%s\n", strings.TrimSpace(recorder.Body.String()))
	if expected, got := wantCode, gotResp.StatusCode; expected != got {
		t.Errorf("expected code %v, got %v", expected, got)
	}

	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	var gotBuf bytes.Buffer
	if err := m.Minify("text/html", &gotBuf, recorder.Body); err != nil {
		panic(err)
	}
	gotBody := gotBuf.String()

	var wantBuf bytes.Buffer
	if err := m.Minify("text/html", &wantBuf, bytes.NewReader(wantBytes)); err != nil {
		panic(err)
	}
	wantBody := wantBuf.String()
	assert.Equal(t, wantBody, gotBody, "body")
}
