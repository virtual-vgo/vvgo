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
	"strings"
	"testing"
	"time"
)

func TestLoginView_ServeHTTP(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		server := LoginView{Sessions: newSessions()}

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		server.ServeHTTP(recorder, request)
		gotResp := recorder.Result()
		assert.Equal(t, http.StatusOK, gotResp.StatusCode)
		wantRaw, gotRaw := strings.TrimSpace(mustReadFile(t, "testdata/login.html")), strings.TrimSpace(recorder.Body.String())
		assertEqualHTML(t, wantRaw, gotRaw)
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
		assertEqualHTML(t, "<a href=/>Found</a>.", strings.TrimSpace(recorder.Body.String()))
	})
}

func TestPartsView_ServeHTTP(t *testing.T) {
	ctx := context.Background()
	handlerStorage := Database{
		Parts:  newParts(),
		Distro: &storage.Bucket{Name: "vvgo-distro"},
	}

	// load the cache with some dummy data
	require.NoError(t, handlerStorage.Parts.Save(ctx, []parts.Part{
		{
			ID: parts.ID{
				Project: "01-snake-eater",
				Name:    "trumpet 3",
			},
			Sheets: []parts.Link{{ObjectKey: "sheet.pdf", CreatedAt: time.Now()}},
			Clix:   []parts.Link{{ObjectKey: "click.mp3", CreatedAt: time.Now()}},
		},
		{
			ID: parts.ID{
				Project: "02-proof-of-a-hero",
				Name:    "trumpet 3",
			},
			Sheets: []parts.Link{{ObjectKey: "sheet.pdf", CreatedAt: time.Now()}},
			Clix:   []parts.Link{{ObjectKey: "click.mp3", CreatedAt: time.Now()}},
		},
		{
			ID: parts.ID{
				Project: "03-the-end-begins-to-rock",
				Name:    "trumpet 3",
			},
			Sheets: []parts.Link{{ObjectKey: "sheet.pdf", CreatedAt: time.Now()}},
			Clix:   []parts.Link{{ObjectKey: "click.mp3", CreatedAt: time.Now()}},
		},
	}), "parts.Save()")

	server := PartView{&handlerStorage}

	t.Run("accept:application/json", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/sheets", nil)
		request.Header.Set("Accept", "application/json")
		server.ServeHTTP(recorder, request)

		wantRaw, gotRaw := strings.TrimSpace(mustReadFile(t, "testdata/parts.json")), strings.TrimSpace(recorder.Body.String())
		var wantMap []map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(wantRaw), &wantMap), "json.Unmarshal")
		var gotMap []map[string]interface{}
		assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &gotMap), "json.Unmarshal")
		if !assert.Equal(t, wantMap, gotMap, "body") {
			t.Logf("Expected body:\n%s\n", wantRaw)
			t.Logf("Got body:\n%s\n", gotRaw)
		}
	})

	t.Run("accept:text/html", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/sheets", nil)
		request.Header.Set("Accept", "text/html")
		server.ServeHTTP(recorder, request)
		got := recorder.Result()
		assert.Equal(t, http.StatusOK, got.StatusCode)
		assertEqualHTML(t, mustReadFile(t, "testdata/parts.html"), recorder.Body.String())
	})
}

func TestIndexView_ServeHTTP(t *testing.T) {
	server := IndexView{}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	server.ServeHTTP(recorder, request)
	got := recorder.Result()
	assert.Equal(t, http.StatusOK, got.StatusCode)
	assertEqualHTML(t, mustReadFile(t, "testdata/index.html"), recorder.Body.String())
}

func mustReadFile(t *testing.T, fileName string) string {
	wantBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	return string(wantBytes)
}

func assertEqualHTML(t *testing.T, want string, got string) {
	want = strings.TrimSpace(want)
	got = strings.TrimSpace(got)
	gotRaw := got + "\n"
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	var gotBuf bytes.Buffer
	if err := m.Minify("text/html", &gotBuf, strings.NewReader(got)); err != nil {
		panic(err)
	}
	gotBody := gotBuf.String()

	var wantBuf bytes.Buffer
	if err := m.Minify("text/html", &wantBuf, strings.NewReader(want)); err != nil {
		panic(err)
	}
	wantBody := wantBuf.String()
	if !assert.Equal(t, wantBody, gotBody, "body") {
		t.Logf("Got Body:\n%s\n", strings.TrimSpace(gotRaw))
	}
}
