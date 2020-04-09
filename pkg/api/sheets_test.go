package api

import (
	"bytes"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"golang.org/x/net/html"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApiServer_SheetsIndex(t *testing.T) {
	type wants struct {
		code int
		body string
	}

	for _, tt := range []struct {
		name        string
		objectStore MockObjectStore
		request     *http.Request
		wants       wants
	}{
		{
			name:    "method post",
			request: httptest.NewRequest(http.MethodPost, "/sheets", nil),
			wants:   wants{code: http.StatusMethodNotAllowed},
		},
		{
			name: "method get",
			objectStore: MockObjectStore{
				listObjects: func(bucketName string) []storage.Object {
					return []storage.Object{
						{
							ContentType: "application/pdf",
							Name:        "midnight-trumpet-3.pdf",
							Tags: map[string]string{
								"Project":     "midnight",
								"Instrument":  "trumpet",
								"Part-Number": "3",
							},
						},
						{
							ContentType: "application/pdf",
							Name:        "daylight-flute-2.pdf",
							Tags: map[string]string{
								"Project":     "daylight",
								"Instrument":  "flute",
								"Part-Number": "2",
							},
						},
					}
				},
			},
			request: httptest.NewRequest(http.MethodGet, "/sheets", strings.NewReader("")),
			wants: wants{
				code: http.StatusOK,
				body: `[{"project":"midnight","instrument":"trumpet","part_number":3,"link":"/download?bucket=sheets\u0026key=midnight-trumpet-3.pdf"},{"project":"daylight","instrument":"flute","part_number":2,"link":"/download?bucket=sheets\u0026key=daylight-flute-2.pdf"}]`,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			apiServer := NewServer(tt.objectStore, Config{MaxContentLength: 1e3})
			recorder := httptest.NewRecorder()
			apiServer.ServeHTTP(recorder, tt.request)
			gotResp := recorder.Result()
			gotBody := recorder.Body.String()
			if expected, got := tt.wants.code, gotResp.StatusCode; expected != got {
				t.Errorf("expected code %v, got %v", expected, got)
			}
			if expected, got := tt.wants.body, strings.TrimSpace(string(gotBody)); expected != got {
				t.Errorf("expected body:\nwant: `%s`\n got: `%s`", expected, got)
			}
		})
	}
}

func TestApiServer_SheetsUpload(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/sheets/upload", strings.NewReader(""))
		wantCode := http.StatusOK
		wantHTML := tokenizeHTMLFile("testdata/test-get-sheets-upload.html")

		apiServer := NewServer(MockObjectStore{}, Config{})
		recorder := httptest.NewRecorder()
		apiServer.ServeHTTP(recorder, request)
		gotResp := recorder.Result()
		gotHTML := html.NewTokenizer(gotResp.Body)

		if expected, got := wantCode, gotResp.StatusCode; expected != got {
			t.Errorf("expected code %v, got %v", expected, got)
		}

		var expected string
		for token := wantHTML.Next(); token != html.ErrorToken; token = wantHTML.Next() {
			expected += string(wantHTML.Raw())
		}

		var got string
		for token := gotHTML.Next(); token != html.ErrorToken; token = gotHTML.Next() {
			got += string(gotHTML.Raw())
		}

		if expected != got {
			t.Errorf("\nwant: `%#v`\n got: `%#v`", expected, got)
		}
	})

	t.Run("post", func(t *testing.T) {
		maxContentLength := int64(1024 * 1024)
		type storeParams struct {
			bucketName string
			object     storage.Object
		}

		type wants struct {
			code        int
			body        string
			storeParams storeParams
		}

		for _, tt := range []struct {
			name        string
			objectStore MockObjectStore
			request     *http.Request
			wants       wants
		}{
			{
				// we should check too many bytes before anything else, so other fields in this mock request
				// are also invalid
				name: "too many bytes",
				request: newPostRequest("/sheets/upload",
					"application/xml", bytes.NewReader(make([]byte, maxContentLength+1))),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return nil }},
				wants: wants{
					code: http.StatusRequestEntityTooLarge,
				},
			},
			{
				name: "invalid content-type",
				request: newPostRequest("/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4",
					"application/xml", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return nil }},
				wants: wants{
					code: http.StatusUnsupportedMediaType,
				},
			},
			{
				name: "multipart/form-data/not a form",
				request: newPostRequest("/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4",
					"multipart/form-data", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return nil }},
				wants: wants{
					code: http.StatusBadRequest,
				},
			},
			{
				name: "multipart/form-data/missing fields",
				request: newFileUploadRequest("/sheets/upload",
					map[string]string{"project": "01-snake-eater", "instrument": "trumpet"},
					"upload_file", "upload.pdf", "application/pdf", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return nil }},
				wants: wants{
					code: http.StatusBadRequest,
					body: sheets.ErrMissingPartNumber.Error(),
				},
			},
			{
				name: "multipart/form-data/file is not a pdf",
				request: newFileUploadRequest("/sheets/upload",
					map[string]string{"project": "01-snake-eater", "instrument": "trumpet", "part_number": "4"},
					"upload_file", "upload.pdf", "application/xml", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return nil }},
				wants: wants{
					code: http.StatusUnsupportedMediaType,
				},
			},
			{
				name: "multipart/form-data/file does not contain pdf data",
				request: newFileUploadRequest("/sheets/upload",
					map[string]string{"project": "01-snake-eater", "instrument": "trumpet", "part_number": "4"},
					"upload_file", "upload.pdf", "application/pdf", strings.NewReader("")),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return nil }},
				wants: wants{
					code: http.StatusUnsupportedMediaType,
				},
			},
			{
				name: "multipart/form-data/db error",
				request: newFileUploadRequest("/sheets/upload",
					map[string]string{"project": "01-snake-eater", "instrument": "trumpet", "part_number": "4"},
					"upload_file", "upload.pdf", "application/pdf", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return fmt.Errorf("mock error") }},
				wants: wants{
					code: http.StatusInternalServerError,
					storeParams: storeParams{
						bucketName: sheets.BucketName,
						object: storage.Object{
							ContentType: "application/pdf",
							Name:        "01-snake-eater-trumpet-4.pdf",
							Tags:        map[string]string{"Project": "01-snake-eater", "Instrument": "trumpet", "Part-Number": "4"},
							Buffer:      *bytes.NewBuffer(mustReadFile("testdata/empty.pdf")),
						},
					},
				},
			},
			{
				name: "multipart/form-data/success",
				request: newFileUploadRequest("/sheets/upload",
					map[string]string{"project": "01-snake-eater", "instrument": "trumpet", "part_number": "4"},
					"upload_file", "upload.pdf", "application/pdf", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return nil }},
				wants: wants{
					code: http.StatusOK,
					storeParams: storeParams{
						bucketName: sheets.BucketName,
						object: storage.Object{
							ContentType: "application/pdf",
							Name:        "01-snake-eater-trumpet-4.pdf",
							Tags:        map[string]string{"Project": "01-snake-eater", "Instrument": "trumpet", "Part-Number": "4"},
							Buffer:      *bytes.NewBuffer(mustReadFile("testdata/empty.pdf")),
						},
					},
				},
			},
			{
				name: "application/pdf/missing fields",
				request: newPostRequest("/sheets/upload?project=01-snake-eater&instrument=trumpet",
					"application/pdf", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return nil }},
				wants: wants{
					code: http.StatusBadRequest,
					body: sheets.ErrMissingPartNumber.Error(),
				},
			},
			{
				name: "application/pdf/body does not contain pdf data",
				request: newPostRequest("/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4",
					"application/pdf", strings.NewReader("")),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return nil }},
				wants: wants{
					code: http.StatusUnsupportedMediaType,
				},
			},
			{
				name: "application/pdf/db error",
				request: newPostRequest("/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4",
					"application/pdf", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return fmt.Errorf("mock error") }},
				wants: wants{
					code: http.StatusInternalServerError,
					storeParams: storeParams{
						bucketName: sheets.BucketName,
						object: storage.Object{
							ContentType: "application/pdf",
							Name:        "01-snake-eater-trumpet-4.pdf",
							Tags:        map[string]string{"Project": "01-snake-eater", "Instrument": "trumpet", "Part-Number": "4"},
							Buffer:      *bytes.NewBuffer(mustReadFile("testdata/empty.pdf")),
						},
					},
				},
			},
			{
				name: "application/pdf/success",
				request: newPostRequest("/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4",
					"application/pdf", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *storage.Object) error { return nil }},
				wants: wants{
					code: http.StatusOK,
					storeParams: storeParams{
						bucketName: sheets.BucketName,
						object: storage.Object{
							ContentType: "application/pdf",
							Name:        "01-snake-eater-trumpet-4.pdf",
							Tags:        map[string]string{"Project": "01-snake-eater", "Instrument": "trumpet", "Part-Number": "4"},
							Buffer:      *bytes.NewBuffer(mustReadFile("testdata/empty.pdf")),
						},
					},
				},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				var gotObject storage.Object
				var gotBucketName string
				apiServer := NewServer(MockObjectStore{
					listObjects: tt.objectStore.listObjects,
					putObject: func(bucketName string, object *storage.Object) error {
						gotObject = *object
						gotBucketName = bucketName
						return tt.objectStore.putObject(bucketName, object)
					},
				}, Config{
					MaxContentLength: maxContentLength,
				})
				recorder := httptest.NewRecorder()
				apiServer.ServeHTTP(recorder, tt.request)
				gotResp := recorder.Result()
				gotBody := recorder.Body.String()
				if expected, got := tt.wants.code, gotResp.StatusCode; expected != got {
					t.Errorf("expected code %v, got %v", expected, got)
				}
				if expected, got := tt.wants.body, strings.TrimSpace(string(gotBody)); expected != got {
					t.Errorf("expected body:\nwant: `%s`\n got: `%s`", expected, got)
				}
				if expected, got := fmt.Sprintf("%#v", tt.wants.storeParams.bucketName), fmt.Sprintf("%#v", gotBucketName); expected != got {
					t.Errorf("\nwant bucket:%v\n got bucket:%v", expected, got)
				}
				if expected, got := fmt.Sprintf("%#v", tt.wants.storeParams.object), fmt.Sprintf("%#v", gotObject); expected != got {
					t.Errorf("\nwant object:%v\n got body:%v", expected, got)
				}
			})
		}
	})

}
