package main

import (
	"bytes"
	"fmt"
	"github.com/minio/minio-go/v6"
	"net/http"
	"net/http/httptest"
	"net/url"
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
				listObjects: func(bucketName string) []Object {
					return []Object{
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
			apiServer := NewApiServer(tt.objectStore, ApiServerConfig{1e3})
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

func Test_acceptsType(t *testing.T) {
	for _, tt := range []struct {
		name   string
		header http.Header
		arg    string
		want   bool
	}{
		{
			name: "yep",
			header: http.Header{
				"Accept": []string{"text/html,application/json,application/pdf", "application/xml,cheese/sandwich"},
			},
			arg:  "application/xml",
			want: true,
		},
		{
			name: "nope",
			header: http.Header{
				"Accept": []string{"text/html,application/json,application/pdf", "application/xml,cheese/sandwich"},
			},
			arg: "sour/cream", want: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if expected, got := tt.want, acceptsType(&http.Request{Header: tt.header}, tt.arg); expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}

func TestApiServer_SheetsUpload(t *testing.T) {
	maxContentLength := int64(1e3)
	type wants struct {
		code       int
		body       string
		bucketName string
		object     Object
	}
	for _, tt := range []struct {
		name        string
		objectStore MockObjectStore
		request     *http.Request
		contentType string
		wants       wants
	}{
		{
			name:        "get",
			request:     httptest.NewRequest(http.MethodGet, "/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4", strings.NewReader("")),
			contentType: "application/pdf",
			wants:       wants{code: http.StatusMethodNotAllowed},
		},
		{
			name:        "body too large",
			request:     httptest.NewRequest(http.MethodPost, "/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4", bytes.NewReader(make([]byte, maxContentLength+1))),
			contentType: "application/pdf",
			wants:       wants{code: http.StatusRequestEntityTooLarge},
		},
		{
			name:        "wrong content type",
			request:     httptest.NewRequest(http.MethodPost, "/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4", strings.NewReader("")),
			contentType: "application/cheese",
			wants:       wants{code: http.StatusUnsupportedMediaType},
		},
		{
			name:        "missing fields",
			request:     httptest.NewRequest(http.MethodPost, "/sheets/upload?project=test-project&instrument=test-instrument", strings.NewReader("")),
			contentType: "application/pdf",
			wants: wants{
				code: http.StatusBadRequest,
				body: ErrMissingPartNumber.Error(),
			},
		},
		{
			name:        "db error",
			request:     httptest.NewRequest(http.MethodPost, "/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4", strings.NewReader(":wave:")),
			contentType: "application/pdf",
			objectStore: MockObjectStore{
				putObject: func(string, *Object) error {
					return fmt.Errorf("mock error")
				},
			},
			wants: wants{
				object: Object{
					ContentType: "application/pdf",
					Name:        "01-snake-eater-trumpet-4.pdf",
					Tags: map[string]string{
						"Project":     "01-snake-eater",
						"Instrument":  "trumpet",
						"Part-Number": "4",
					},
					Buffer: *bytes.NewBufferString(":wave:"),
				},
				bucketName: SheetsBucketName,
				code:       http.StatusInternalServerError,
			},
		},
		{
			name:        "success",
			request:     httptest.NewRequest(http.MethodPost, "/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4", strings.NewReader(":wave:")),
			contentType: "application/pdf",
			objectStore: MockObjectStore{
				putObject: func(string, *Object) error {
					return nil
				},
			},
			wants: wants{
				object: Object{
					ContentType: "application/pdf",
					Name:        "01-snake-eater-trumpet-4.pdf",
					Tags: map[string]string{
						"Project":     "01-snake-eater",
						"Instrument":  "trumpet",
						"Part-Number": "4",
					},
					Buffer: *bytes.NewBufferString(":wave:"),
				},
				bucketName: SheetsBucketName,
				code:       http.StatusOK,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var gotObject Object
			var gotBucketName string
			apiServer := NewApiServer(MockObjectStore{
				listObjects: tt.objectStore.listObjects,
				putObject: func(bucketName string, object *Object) error {
					gotObject = *object
					gotBucketName = bucketName
					return tt.objectStore.putObject(bucketName, object)
				},
			}, ApiServerConfig{
				MaxContentLength: 1e3,
			})
			tt.request.Header.Set("Content-Type", tt.contentType)
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
			if expected, got := fmt.Sprintf("%#v", tt.wants.bucketName), fmt.Sprintf("%#v", gotBucketName); expected != got {
				t.Errorf("\nwant bucket:%v\n got bucket:%v", expected, got)
			}
			if expected, got := fmt.Sprintf("%#v", tt.wants.object), fmt.Sprintf("%#v", gotObject); expected != got {
				t.Errorf("\nwant object:%v\n got body:%v", expected, got)
			}
		})
	}
}

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
			server := NewApiServer(tt.objectStore, ApiServerConfig{})
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

func TestMusicPDFMeta_ToMap(t *testing.T) {
	meta := Sheet{
		Project:    "01-snake-eater",
		Instrument: "trumpet",
		PartNumber: 4,
	}

	wantMap := map[string]string{
		"Project":     "01-snake-eater",
		"Instrument":  "trumpet",
		"Part-Number": "4",
	}
	gotMap := meta.ToTags()
	if expected, got := fmt.Sprintf("%#v", wantMap), fmt.Sprintf("%#v", gotMap); expected != got {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestNewMusicPDFMetaFromTags(t *testing.T) {
	tags := map[string]string{
		"Project":     "01-snake-eater",
		"Instrument":  "trumpet",
		"Part-Number": "4",
	}

	expectedMeta := Sheet{
		Project:    "01-snake-eater",
		Instrument: "trumpet",
		PartNumber: 4,
	}

	gotMeta := NewSheetFromTags(tags)
	if expected, got := fmt.Sprintf("%#v", expectedMeta), fmt.Sprintf("%#v", gotMeta); expected != got {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestMusicPDFMeta_ReadFromUrlValues(t *testing.T) {
	values, err := url.ParseQuery(`project=test-project&instrument=test-instrument&part_number=4`)
	if err != nil {
		t.Fatalf("url.ParseQuery() failed: %v", err)
	}

	expectedMeta := Sheet{
		Project:    "test-project",
		Instrument: "test-instrument",
		PartNumber: 4,
	}

	gotMeta := NewSheetFromUrlValues(values)
	if expected, got := fmt.Sprintf("%#v", expectedMeta), fmt.Sprintf("%#v", gotMeta); expected != got {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestMusicPDFMeta_Validate(t *testing.T) {
	type fields struct {
		Project    string
		Instrument string
		PartNumber int
	}
	for _, tt := range []struct {
		name   string
		fields fields
		want   error
	}{
		{
			name: "valid",
			fields: fields{
				Project:    "test-project",
				Instrument: "test-instrument",
				PartNumber: 6,
			},
			want: nil,
		},
		{
			name: "missing project",
			fields: fields{
				Instrument: "test-instrument",
				PartNumber: 6,
			},
			want: ErrMissingProject,
		},
		{
			name: "missing instrument",
			fields: fields{
				Project:    "test-project",
				PartNumber: 6,
			},
			want: ErrMissingInstrument,
		},
		{
			name: "missing part number",
			fields: fields{
				Project:    "test-project",
				Instrument: "test-instrument",
			},
			want: ErrMissingPartNumber,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			x := &Sheet{
				Project:    tt.fields.Project,
				Instrument: tt.fields.Instrument,
				PartNumber: tt.fields.PartNumber,
			}
			if expected, got := tt.want, x.Validate(); expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}
