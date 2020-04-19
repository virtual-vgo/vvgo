package api

import (
	"bytes"
	"context"
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
	"os"
	"strings"
	"testing"
	"time"
)

func TestPartsHandler_ServeHTTP(t *testing.T) {
	clixBucket := "clix"
	sheetsBucket := "sheets"
	mockBucket := MockBucket{getObject: func(ctx context.Context, name string, dest *storage.Object) bool {
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

	server := PartsHandler{NavBar{}, &Storage{
		Parts:  parts.Parts{Bucket: &mockBucket},
		Sheets: &mockBucket,
		Clix:   &mockBucket,
		ServerConfig: ServerConfig{
			SheetsBucketName: sheetsBucket,
			ClixBucketName:   clixBucket,
		},
	}}

	t.Run("accept:application/json", func(t *testing.T) {
		var wantBody bytes.Buffer
		file, err := os.Open("testdata/parts.json")
		require.NoError(t, err, "os.Open")
		_, err = wantBody.ReadFrom(file)
		require.NoError(t, err, "file.Read")

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/sheets", nil)
		request.Header.Set("Accept", "application/json")
		server.ServeHTTP(recorder, request)

		wantRaw, gotRaw := strings.TrimSpace(wantBody.String()), strings.TrimSpace(recorder.Body.String())
		var wantMap []map[string]interface{}
		err = json.Unmarshal(wantBody.Bytes(), &wantMap)
		require.NoError(t, err, "json.Unmarshal")
		var gotMap []map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &gotMap)
		assert.NoError(t, err, "json.Unmarshal")
		if !assert.Equal(t, wantMap, gotMap, "body") {
			t.Logf("Expected body:\n%s\n", wantRaw)
			t.Logf("Got body:\n%s\n", gotRaw)
		}
	})

	t.Run("accept:text/html", func(t *testing.T) {
		var wantBody bytes.Buffer
		file, err := os.Open("testdata/parts.html")
		require.NoError(t, err, "os.Open")
		_, err = wantBody.ReadFrom(file)
		require.NoError(t, err, "file.Read")

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/sheets", nil)
		request.Header.Set("Accept", "text/html")
		server.ServeHTTP(recorder, request)

		wantRaw, gotRaw := strings.TrimSpace(wantBody.String()), strings.TrimSpace(recorder.Body.String())
		m := minify.New()
		m.AddFunc("text/html", html.Minify)
		var wantMin bytes.Buffer
		err = m.Minify("text/html", &wantMin, &wantBody)
		require.NoError(t, err, "m.Minify")
		var gotMin bytes.Buffer
		err = m.Minify("text/html", &gotMin, recorder.Body)
		assert.NoError(t, err, "m.Minify")
		if !assert.Equal(t, wantMin.String(), gotMin.String(), "body") {
			t.Logf("Expected body:\n%s\n", wantRaw)
			t.Logf("Got body:\n%s\n", gotRaw)
		}
	})
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
	if !assert.Equal(t, wantBody, gotBody, "body") {
		t.Logf("Got Body:\n%s\n", strings.TrimSpace(recorder.Body.String()))
	}
}
