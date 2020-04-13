package api

import (
	"bytes"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestClient_Upload(t *testing.T) {
	wantUser := "dio"
	wantPass := "brando"
	wantURI := "/upload"
	wantBody := ``
	wantMethod := http.MethodPost
	wantContentType := "application/octet-stream"
	wantStatuses := []UploadStatus{{
		FileName: "Dio_Brando.pdf",
		Code:     http.StatusOK,
	}}
	wantErr := error(nil)

	client := NewClient(ClientConfig{
		BasicAuthUser: "dio",
		BasicAuthPass: "brando",
	})

	var gotRequest *http.Request
	var gotUser, gotPass, gotURI string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r
		gotUser, gotPass, _ = r.BasicAuth()
		gotURI = r.URL.RequestURI()
		gob.NewEncoder(w).Encode([]UploadStatus{{
			FileName: "Dio_Brando.pdf",
			Code:     http.StatusOK,
		}})
	}))
	defer ts.Close()
	client.ServerAddress = ts.URL

	fileBytes, err := ioutil.ReadFile("testdata/sheet-music.pdf")
	if err != nil {
		t.Fatalf("ioutil.ReadAll() failed: %v", err)
	}

	uploads := []Upload{{
		UploadType:  UploadTypeSheets,
		PartNames:   []string{"trumpet"},
		PartNumbers: []uint8{2},
		Project:     "01-snake-eater",
		FileName:    "Dio_Brando.pdf",
		FileBytes:   fileBytes,
		ContentType: "application/pdf",
	}}

	gotStatuses, gotErr := client.Upload(uploads...)

	if want, got := wantURI, gotURI; want != got {
		t.Errorf("expected user `%s`, got `%s`", want, got)
	}
	if want, got := wantUser, gotUser; want != got {
		t.Errorf("expected user `%s`, got `%s`", want, got)
	}
	if want, got := wantPass, gotPass; want != got {
		t.Errorf("expected pass `%s`, got `%s`", want, got)
	}
	if want, got := wantStatuses, gotStatuses; !reflect.DeepEqual(want, got) {
		t.Errorf("expected statuses %#v, got %#v", want, got)
	}
	if want, got := wantErr, gotErr; want != got {
		t.Errorf("expected error %#v, got %#v", want, got)
	}
	if want, got := wantContentType, gotRequest.Header.Get("Content-Type"); want != got {
		t.Errorf("expected content-type `%s`, got `%s`", want, got)
	}
	if want, got := wantMethod, gotRequest.Method; want != got {
		t.Errorf("expected method `%s`, got `%s`", want, got)
	}
	var gotBody bytes.Buffer
	gotBody.ReadFrom(gotRequest.Body)
	if want, got := wantBody, gotBody.String(); want != got {
		t.Errorf("expected body `%s`, got `%s`", want, got)
	}
}

func TestClient_Authenticate(t *testing.T) {
	type wants struct {
		user  string
		pass  string
		uri   string
		error bool
	}

	for _, tt := range []struct {
		name   string
		client *Client
		code   int
		wants  wants
	}{
		{
			name: "success",
			client: NewClient(ClientConfig{
				BasicAuthUser: "dio",
				BasicAuthPass: "brando",
			}),
			code: http.StatusOK,
			wants: wants{
				user:  "dio",
				pass:  "brando",
				uri:   "/auth",
				error: false,
			},
		},
		{
			name: "failure",
			client: NewClient(ClientConfig{
				BasicAuthUser: "dio",
				BasicAuthPass: "brando",
			}),
			code: http.StatusUnauthorized,
			wants: wants{
				user:  "dio",
				pass:  "brando",
				uri:   "/auth",
				error: true,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var gotUser, gotPass, gotURI string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUser, gotPass, _ = r.BasicAuth()
				gotURI = r.URL.RequestURI()
				w.WriteHeader(tt.code)
			}))
			defer ts.Close()
			tt.client.ServerAddress = ts.URL

			gotErr := tt.client.Authenticate()

			assert.Equal(t, tt.wants.uri, gotURI, "uri")
			assert.Equal(t, tt.wants.user, gotUser, "user")
			assert.Equal(t, tt.wants.pass, gotPass, "pass")

			if tt.wants.error {
				assert.Error(t, gotErr, "error")
			} else {
				assert.NoError(t, gotErr, "error")
			}
		})
	}
}
