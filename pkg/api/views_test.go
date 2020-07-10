package api

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/virtual-vgo/vvgo/pkg/login"
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
		loginView := LoginView{Sessions: newSessions()}
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loginView.ServeHTTP(w, r.Clone(context.WithValue(ctx, CtxKeyVVGOIdentity, &login.Identity{Roles: []login.Role{login.RoleVVGOMember}})))
		}))
		defer ts.Close()

		cookie, err := loginView.Sessions.NewCookie(ctx, &login.Identity{Roles: []login.Role{login.RoleVVGOMember}}, 600*time.Second)
		require.NoError(t, err, "sessions.NewCookie()")

		req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
		require.NoError(t, err, "http.NewRequest()")
		req.AddCookie(cookie)
		resp, err := noFollow(nil).Do(req)
		require.NoError(t, err, "http.Do()")
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/login/success", resp.Header.Get("Location"))
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
