package api

import (
	"fmt"
	"github.com/minio/minio-go/v6"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApiServer_Download(t *testing.T) {
	type wants struct {
		code     int
		location string
		body     string
	}

	for _, tt := range []struct {
		name        string
		objectStore MockObjectStore
		request     *http.Request
		wants       wants
	}{
		{
			name: "post",
			objectStore: MockObjectStore{downloadURL: func(bucketName string, objectName string) (string, error) {
				return fmt.Sprintf("http://storage.example.com/%s/%s", bucketName, objectName), nil
			}},
			request: httptest.NewRequest(http.MethodPost, "/download?bucket=cheese&key=danish", strings.NewReader("")),
			wants: wants{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "invalid bucket",
			objectStore: MockObjectStore{downloadURL: func(bucketName string, objectName string) (string, error) {
				return "", minio.ErrorResponse{StatusCode: http.StatusNotFound}
			}},
			request: httptest.NewRequest(http.MethodGet, "/download?bucket=cheese&key=danish", strings.NewReader("")),
			wants: wants{
				code: http.StatusNotFound,
				body: "404 page not found",
			},
		},
		{
			name: "invalid object",
			objectStore: MockObjectStore{downloadURL: func(bucketName string, objectName string) (string, error) {
				return "", minio.ErrorResponse{StatusCode: http.StatusNotFound}
			}},
			request: httptest.NewRequest(http.MethodGet, "/download?bucket=cheese&key=danish", strings.NewReader("")),
			wants: wants{
				code: http.StatusNotFound,
				body: "404 page not found",
			},
		},
		{
			name: "server error",
			objectStore: MockObjectStore{downloadURL: func(bucketName string, objectName string) (string, error) {
				return "", fmt.Errorf("mock error")
			}},
			request: httptest.NewRequest(http.MethodGet, "/download?bucket=cheese&key=danish", strings.NewReader("")),
			wants:   wants{code: http.StatusInternalServerError},
		},
		{
			name: "success",
			objectStore: MockObjectStore{downloadURL: func(bucketName string, objectName string) (string, error) {
				return fmt.Sprintf("http://storage.example.com/%s/%s", bucketName, objectName), nil
			}},
			request: httptest.NewRequest(http.MethodGet, "/download?bucket=cheese&key=danish", strings.NewReader("")),
			wants: wants{
				code:     http.StatusFound,
				location: "http://storage.example.com/cheese/danish",
				body:     `<a href="http://storage.example.com/cheese/danish">Found</a>.`,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(tt.objectStore, Config{})
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, tt.request)
			gotResp := recorder.Result()
			if expected, got := tt.wants.code, gotResp.StatusCode; expected != got {
				t.Errorf("expected code %v, got %v", expected, got)
			}
			if expected, got := tt.wants.location, gotResp.Header.Get("Location"); expected != got {
				t.Errorf("expected location %v, got %v", expected, got)
			}
			if expected, got := tt.wants.body, strings.TrimSpace(recorder.Body.String()); expected != got {
				t.Errorf("expected body %v, got %v", expected, got)
			}
		})
	}
}
