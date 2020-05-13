package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/virtual-vgo/vvgo/pkg/login"
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

func TestLoginView_ServeHTTP(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		wantCode := http.StatusOK
		wantBytes, err := ioutil.ReadFile("testdata/login.html")
		if err != nil {
			t.Fatalf("ioutil.ReadFile() failed: %v", err)
		}

		server := LoginView{Sessions: newSessions()}

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		server.ServeHTTP(recorder, request)
		gotResp := recorder.Result()
		if expected, got := wantCode, gotResp.StatusCode; expected != got {
			t.Errorf("expected code %v, got %v", expected, got)
		}
		assertEqualHtml(t, wantBytes, recorder)
	})

	t.Run("logged in", func(t *testing.T) {
		ctx := context.Background()

		server := LoginView{Sessions: newSessions()}

		cookie, err := server.Sessions.NewCookie(ctx, &login.Identity{Roles: []login.Role{login.RoleVVGOMember}}, 600*time.Second)
		require.NoError(t, err, "sessions.NewCookie()")

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.AddCookie(cookie)
		server.ServeHTTP(recorder, request)
		gotResp := recorder.Result()
		if expected, got := http.StatusFound, gotResp.StatusCode; expected != got {
			t.Errorf("expected code %v, got %v", expected, got)
		}
		assertEqualHtml(t, []byte("<a href=/>Found</a>."), recorder)
	})
}

func TestPartsView_ServeHTTP(t *testing.T) {
	warehouse, err := storage.NewWarehouse(storage.Config{NoOp: true})
	require.NoError(t, err, "storage.NewWarehouse")

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

	// load the cache with some dummy data
	handlerStorage.Parts.Save(ctx, []parts.Part{
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
	})

	server := PartView{NavBar{}, &handlerStorage}

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
		assertEqualHtml(t, wantBody.Bytes(), recorder)
	})
}

func TestIndexView_ServeHTTP(t *testing.T) {
	wantCode := http.StatusOK
	wantBytes, err := ioutil.ReadFile("testdata/index.html")
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}

	server := IndexView{}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	server.ServeHTTP(recorder, request)
	gotResp := recorder.Result()
	if expected, got := wantCode, gotResp.StatusCode; expected != got {
		t.Errorf("expected code %v, got %v", expected, got)
	}
	assertEqualHtml(t, wantBytes, recorder)
}

func assertEqualHtml(t *testing.T, wantBytes []byte, recorder *httptest.ResponseRecorder) {
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
