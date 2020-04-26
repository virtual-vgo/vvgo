package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"github.com/virtual-vgo/vvgo/pkg/locker"
	"github.com/virtual-vgo/vvgo/pkg/parts"
	"github.com/virtual-vgo/vvgo/pkg/sessions"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPartsView_ServeHTTP(t *testing.T) {
	warehouse, err := storage.NewWarehouse(storage.Config{})
	require.NoError(t, err, "storage.NewWarehouse")

	ctx := context.Background()
	bucket, err := warehouse.NewBucket(ctx, "testing")
	require.NoError(t, err, "storage.NewBucket")
	handlerStorage := Storage{
		Parts: &parts.Parts{
			Cache:  storage.NewCache(storage.CacheOpts{}),
			Locker: locker.NewLocker(locker.Opts{}),
		},
		Sheets: bucket,
		Clix:   bucket,
		Tracks: bucket,
		StorageConfig: StorageConfig{
			SheetsBucketName: "sheets",
			ClixBucketName:   "clix",
			PartsBucketName:  "parts",
			TracksBucketName: "tracks",
		},
	}

	// load the cache with some dummy data
	obj := storage.Object{ContentType: "application/json"}
	require.NoError(t, json.NewEncoder(&obj.Buffer).Encode([]parts.Part{
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
	}), "json.Encode()")
	require.NoError(t, handlerStorage.Parts.Cache.PutObject(ctx, parts.DataFile, &obj), "cache.PutObject()")

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

func TestLoginView_ServeHTTP(t *testing.T) {
	loginView := LoginView{
		Sessions: sessions.NewStore(sessions.Secret{}, sessions.Config{CookieName: "vvgo-cookie"}),
	}

	t.Run("redirect", func(t *testing.T) {
		ctx := context.Background()
		loginView.Sessions.Init(context.Background())
		ts := httptest.NewServer(loginView)
		defer ts.Close()
		tsUrl, err := url.Parse(ts.URL)
		require.NoError(t, err, "url.Parse()")

		// create a session and cookie
		session := loginView.Sessions.NewSession(time.Now().Add(7 * 24 * 3600 * time.Second))
		cookie := loginView.Sessions.NewCookie(session)
		assert.NoError(t, loginView.Sessions.StoreIdentity(ctx, session.ID, &sessions.Identity{
			Kind:  sessions.KindPassword,
			Roles: []sessions.Role{"cheese"},
		}))

		// set the cookie on the client
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		require.NoError(t, err, "cookiejar.New")
		jar.SetCookies(tsUrl, []*http.Cookie{cookie})

		client := noFollow(&http.Client{Jar: jar})
		resp, err := client.Get(ts.URL)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/", resp.Header.Get("Location"), "location")
	})

	t.Run("view", func(t *testing.T) {
		wantCode := http.StatusOK
		wantBytes, err := ioutil.ReadFile("testdata/login.html")
		if err != nil {
			t.Fatalf("ioutil.ReadFile() failed: %v", err)
		}

		loginView.Sessions.Init(context.Background())
		ts := httptest.NewServer(loginView)
		defer ts.Close()

		require.NoError(t, err, "cookiejar.New")
		resp, err := noFollow(http.DefaultClient).Get(ts.URL)
		require.NoError(t, err, "client.Get")
		assert.Equal(t, wantCode, resp.StatusCode)
		var respBody bytes.Buffer
		_, err = respBody.ReadFrom(resp.Body)
		require.NoError(t, err, "resp.Body.Read() failed")
		origBody := strings.TrimSpace(respBody.String())

		m := minify.New()
		m.AddFunc("text/html", html.Minify)
		var gotBuf bytes.Buffer
		if err := m.Minify("text/html", &gotBuf, &respBody); err != nil {
			panic(err)
		}
		gotBody := gotBuf.String()

		var wantBuf bytes.Buffer
		if err := m.Minify("text/html", &wantBuf, bytes.NewReader(wantBytes)); err != nil {
			panic(err)
		}
		wantBody := wantBuf.String()
		if !assert.Equal(t, wantBody, gotBody, "body") {
			t.Logf("Got Body:\n%s\n", origBody)
		}
	})
}
