package api

import (
	"context"
	"encoding/gob"
	"github.com/stretchr/testify/assert"
	"github.com/virtual-vgo/vvgo/pkg/login"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestClient_Backup(t *testing.T) {
	server := NewServer(context.Background(), ServerConfig{
		MemberUser:        "vvgo-member",
		MemberPass:        "vvgo-member",
		UploaderToken:     "vvgo-uploader",
		DeveloperToken:    "vvgo-developer",
		DistroBucketName:  "vvgo-distro" + strconv.Itoa(lrand.Int()),
		BackupsBucketName: "vvgo-backups" + strconv.Itoa(lrand.Int()),
		RedisNamespace:    "vvgo-testing" + strconv.Itoa(lrand.Int()),
		Login: login.Config{
			CookieName: "vvgo-cookie",
		},
	})
	ts := httptest.NewServer(http.HandlerFunc(server.Server.Handler.ServeHTTP))
	defer ts.Close()

	client := NewAsyncClient(AsyncClientConfig{
		ClientConfig: ClientConfig{
			ServerAddress: ts.URL,
			Token:         "vvgo-uploader",
		},
	})
	assert.NoError(t, client.Backup())
}

func TestClient_GetProject(t *testing.T) {
	ts := httptest.NewServer(ProjectsHandler{})
	defer ts.Close()
	client := NewAsyncClient(AsyncClientConfig{
		ClientConfig: ClientConfig{
			Token:         "Dio Brando",
			ServerAddress: ts.URL,
		},
		MaxParallel: 32,
		QueueLength: 64,
	})
	got, err := client.GetProject("01-snake-eater")
	assert.NoError(t, err)
	assert.Equal(t, projects.GetName("01-snake-eater"), got)
}

func TestClient_Upload(t *testing.T) {
	var gotRequest *http.Request
	var gotAuthorized bool
	mux := RBACMux{
		Bearer:   map[string][]login.Role{"Dio Brando": {login.RoleVVGOUploader}},
		ServeMux: http.NewServeMux(),
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gotRequest = r.Clone(context.Background())
		gotAuthorized = true
		gob.NewEncoder(w).Encode([]UploadStatus{{
			FileName: "Dio_Brando.pdf",
			Code:     http.StatusOK,
		}})
	}, login.RoleVVGOUploader)
	ts := httptest.NewServer(&mux)
	defer ts.Close()

	client := NewAsyncClient(AsyncClientConfig{
		ClientConfig: ClientConfig{
			Token:         "Dio Brando",
			ServerAddress: ts.URL,
		},
		MaxParallel: 32,
		QueueLength: 64,
	})

	fileBytes, err := ioutil.ReadFile("testdata/sheet-music.pdf")
	if err != nil {
		t.Fatalf("ioutil.ReadAll() failed: %v", err)
	}

	uploads := []Upload{{
		UploadType:  UploadTypeSheets,
		PartNames:   []string{"trumpet 2"},
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

	assert.Equal(t, "/upload", gotRequest.URL.RequestURI(), "request uri")
	assert.Equal(t, []UploadStatus{{
		FileName: "Dio_Brando.pdf",
		Code:     http.StatusOK,
	}}, gotStatuses, "upload status")
	assert.True(t, gotAuthorized, "authorized")
	assert.Equal(t, MediaTypeUploadsGob, gotRequest.Header.Get("Content-Type"), "content-type")
	assert.Equal(t, "application/gzip", gotRequest.Header.Get("Content-Encoding"), "content-encoding")
	assert.Equal(t, http.MethodPost, gotRequest.Method, "method")
}
