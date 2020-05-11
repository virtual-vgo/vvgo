package api

import (
	"context"
	"encoding/gob"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestClient_Upload(t *testing.T) {
	wantURI := "/upload"
	wantMethod := http.MethodPost
	wantContentType := MediaTypeUploadsGob
	wantContentEncoding := "application/gzip"
	wantAuthorized := true
	wantStatuses := []UploadStatus{{
		FileName: "Dio_Brando.pdf",
		Code:     http.StatusOK,
	}}

	client := NewAsyncClient(AsyncClientConfig{
		ClientConfig: ClientConfig{
			Token: "Dio Brando",
		},
		MaxParallel: 32,
		QueueLength: 64,
	})

	var gotRequest *http.Request
	var gotAuthorized bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r.Clone(context.Background())
		gotAuthorized = true
		gob.NewEncoder(w).Encode([]UploadStatus{{
			FileName: "Dio_Brando.pdf",
			Code:     http.StatusOK,
		}})
	}))
	defer ts.Close()
	client.Client.ServerAddress = ts.URL

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

	client.Upload(uploads...)
	client.Close()
	var gotStatuses []UploadStatus
	for status := range client.Status() {
		gotStatuses = append(gotStatuses, status)
	}

	if want, got := wantURI, gotRequest.URL.RequestURI(); want != got {
		t.Errorf("expected user `%s`, got `%s`", want, got)
	}
	if want, got := wantStatuses, gotStatuses; !reflect.DeepEqual(want, got) {
		t.Errorf("expected statuses %#v, got %#v", want, got)
	}
	if want, got := wantAuthorized, gotAuthorized; want != got {
		t.Errorf("expected authorized `%v`, got `%v`", want, got)
	}
	if want, got := wantContentType, gotRequest.Header.Get("Content-Type"); want != got {
		t.Errorf("expected content-type `%s`, got `%s`", want, got)
	}
	if want, got := wantContentEncoding, gotRequest.Header.Get("Content-Encoding"); want != got {
		t.Errorf("expected content-encoding `%s`, got `%s`", want, got)
	}
	if want, got := wantMethod, gotRequest.Method; want != got {
		t.Errorf("expected method `%s`, got `%s`", want, got)
	}
}
