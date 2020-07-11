package api

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
