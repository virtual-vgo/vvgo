package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/minio/minio-go/v6"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"strings"
	"testing"
)

func tokenizeHTMLFile(src string) *html.Tokenizer {
	file, err := os.Open(src)
	if err != nil {
		panic(fmt.Errorf("os.Open() failed: %v", err))
	}
	return html.NewTokenizer(file)
}

func TestApiServer_Authenticate(t *testing.T) {
	type wants struct {
		code int
		body string
	}

	var newAuthRequest = func(url, user, pass string) *http.Request {
		req := httptest.NewRequest(http.MethodGet, url, strings.NewReader(""))
		req.SetBasicAuth(user, pass)
		return req
	}

	for _, tt := range []struct {
		name    string
		config  ApiServerConfig
		request *http.Request
		wants   wants
	}{
		{
			name:    "success",
			request: newAuthRequest("/", "jackson", "the-earth-is-flat"),
			config:  ApiServerConfig{BasicAuthUser: "jackson", BasicAuthPass: "the-earth-is-flat"},
			wants:   wants{code: http.StatusOK},
		},
		{
			name:    "incorrect user",
			request: newAuthRequest("/", "", "the-earth-is-flat"),
			config:  ApiServerConfig{BasicAuthUser: "jackson", BasicAuthPass: "the-earth-is-flat"},
			wants: wants{
				code: http.StatusUnauthorized,
				body: "authorization failed",
			},
		},
		{
			name:    "incorrect pass",
			request: newAuthRequest("/", "jackson", ""),
			config:  ApiServerConfig{BasicAuthUser: "jackson", BasicAuthPass: "the-earth-is-flat"},
			wants: wants{
				code: http.StatusUnauthorized,
				body: "authorization failed",
			},
		},
		{
			name:    "no auth",
			request: httptest.NewRequest(http.MethodGet, "/", strings.NewReader("")),
			config:  ApiServerConfig{BasicAuthUser: "jackson", BasicAuthPass: "the-earth-is-flat"},
			wants: wants{
				code: http.StatusUnauthorized,
				body: "authorization failed",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server := NewApiServer(MockObjectStore{}, tt.config)
			server.Authenticate(func(w http.ResponseWriter, r *http.Request) {
				// do nothing
			})(recorder, tt.request)

			gotCode := recorder.Code
			gotBody := strings.TrimSpace(recorder.Body.String())

			if expected, got := tt.wants.code, gotCode; expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
			if expected, got := tt.wants.body, gotBody; expected != got {
				t.Errorf("expected %v, got %v", expected, got)
			}
		})
	}
}

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
			apiServer := NewApiServer(tt.objectStore, ApiServerConfig{MaxContentLength: 1e3})
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
	t.Run("get", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/sheets/upload", strings.NewReader(""))
		wantCode := http.StatusOK
		wantHTML := tokenizeHTMLFile("testdata/test-get-sheets-upload.html")

		apiServer := NewApiServer(MockObjectStore{}, ApiServerConfig{})
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
			object     Object
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
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return nil }},
				wants: wants{
					code: http.StatusRequestEntityTooLarge,
				},
			},
			{
				name: "invalid content-type",
				request: newPostRequest("/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4",
					"application/xml", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return nil }},
				wants: wants{
					code: http.StatusUnsupportedMediaType,
				},
			},
			{
				name: "multipart/form-data/not a form",
				request: newPostRequest("/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4",
					"multipart/form-data", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return nil }},
				wants: wants{
					code: http.StatusBadRequest,
				},
			},
			{
				name: "multipart/form-data/missing fields",
				request: newFileUploadRequest("/sheets/upload",
					map[string]string{"project": "01-snake-eater", "instrument": "trumpet"},
					"upload_file", "upload.pdf", "application/pdf", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return nil }},
				wants: wants{
					code: http.StatusBadRequest,
					body: ErrMissingPartNumber.Error(),
				},
			},
			{
				name: "multipart/form-data/file is not a pdf",
				request: newFileUploadRequest("/sheets/upload",
					map[string]string{"project": "01-snake-eater", "instrument": "trumpet", "part_number": "4"},
					"upload_file", "upload.pdf", "application/xml", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return nil }},
				wants: wants{
					code: http.StatusUnsupportedMediaType,
				},
			},
			{
				name: "multipart/form-data/file does not contain pdf data",
				request: newFileUploadRequest("/sheets/upload",
					map[string]string{"project": "01-snake-eater", "instrument": "trumpet", "part_number": "4"},
					"upload_file", "upload.pdf", "application/pdf", strings.NewReader("")),
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return nil }},
				wants: wants{
					code: http.StatusUnsupportedMediaType,
				},
			},
			{
				name: "multipart/form-data/db error",
				request: newFileUploadRequest("/sheets/upload",
					map[string]string{"project": "01-snake-eater", "instrument": "trumpet", "part_number": "4"},
					"upload_file", "upload.pdf", "application/pdf", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return fmt.Errorf("mock error") }},
				wants: wants{
					code: http.StatusInternalServerError,
					storeParams: storeParams{
						bucketName: SheetsBucketName,
						object: Object{
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
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return nil }},
				wants: wants{
					code: http.StatusOK,
					storeParams: storeParams{
						bucketName: SheetsBucketName,
						object: Object{
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
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return nil }},
				wants: wants{
					code: http.StatusBadRequest,
					body: ErrMissingPartNumber.Error(),
				},
			},
			{
				name: "application/pdf/body does not contain pdf data",
				request: newPostRequest("/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4",
					"application/pdf", strings.NewReader("")),
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return nil }},
				wants: wants{
					code: http.StatusUnsupportedMediaType,
				},
			},
			{
				name: "application/pdf/db error",
				request: newPostRequest("/sheets/upload?project=01-snake-eater&instrument=trumpet&part_number=4",
					"application/pdf", bytes.NewReader(mustReadFile("testdata/empty.pdf"))),
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return fmt.Errorf("mock error") }},
				wants: wants{
					code: http.StatusInternalServerError,
					storeParams: storeParams{
						bucketName: SheetsBucketName,
						object: Object{
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
				objectStore: MockObjectStore{putObject: func(string, *Object) error { return nil }},
				wants: wants{
					code: http.StatusOK,
					storeParams: storeParams{
						bucketName: SheetsBucketName,
						object: Object{
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

func newPostRequest(uri string, contentType string, src io.Reader) *http.Request {
	req, err := http.NewRequest(http.MethodPost, uri, src)
	if err != nil {
		panic(fmt.Sprintf("http.NewRequest() failed: %v", err))
	}
	req.Header.Set("Content-Type", contentType)
	return req
}

// Creates a new file upload http request with optional extra params
func newFileUploadRequest(uri string, params map[string]string, fileParam, fileName, contentType string, src io.Reader) *http.Request {
	escapeQuotes := strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace

	if r, err := func() (*http.Request, error) {
		var body bytes.Buffer
		multipartWriter := multipart.NewWriter(&body)

		fileHeader := make(textproto.MIMEHeader)
		fileHeader.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				escapeQuotes(fileParam), escapeQuotes(fileName)))
		fileHeader.Set("Content-Type", contentType)
		fileDest, err := multipartWriter.CreatePart(fileHeader)

		if err != nil {
			return nil, fmt.Errorf("multipartWriter.CreateFormFile() failed: %v", err)
		}

		if _, err = io.Copy(fileDest, src); err != nil {
			return nil, fmt.Errorf("io.Copy() failed: %v", err)
		}

		for key, val := range params {
			if err = multipartWriter.WriteField(key, val); err != nil {
				return nil, fmt.Errorf("multipartWriter.WriteField() failed: %v", err)
			}
		}

		if err = multipartWriter.Close(); err != nil {
			return nil, fmt.Errorf("multipartWriter.Close() failed: %v", err)
		}

		if req, err := http.NewRequest("POST", uri, &body); err != nil {
			return nil, fmt.Errorf("http.NewRequest() failed: %v", err)
		} else {
			req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
			return req, nil
		}
	}(); err != nil {
		panic(err)
	} else {
		return r
	}
}

func mustReadFile(fileName string) []byte {
	if buf, err := ioutil.ReadFile(fileName); err != nil {
		panic(fmt.Sprintf("ioutil.ReadFile() failed: %v", err))
	} else {
		return buf
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

func TestSheet_ToTags(t *testing.T) {
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
	gotMap := meta.Tags()
	if expected, got := fmt.Sprintf("%#v", wantMap), fmt.Sprintf("%#v", gotMap); expected != got {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestApiServer_Version(t *testing.T) {
	t.Run("accept:application/json", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/version", strings.NewReader(""))
		req.Header.Set("Accept", "application/json")
		apiServer := NewApiServer(MockObjectStore{}, ApiServerConfig{})
		apiServer.ServeHTTP(recorder, req)

		if expected, got := http.StatusOK, recorder.Code; expected != got {
			t.Errorf("expected code %v, got %v", expected, got)
		}
		var gotJSON json.RawMessage
		if err := json.NewDecoder(recorder.Body).Decode(&gotJSON); err != nil {
			t.Errorf("json.Decode() failed: %v", err)
		}

	})
	t.Run("accept:text/html", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/version", strings.NewReader(""))
		req.Header.Set("Accept", "text/html")

		apiServer := NewApiServer(MockObjectStore{}, ApiServerConfig{})
		apiServer.ServeHTTP(recorder, req)
		if expected, got := http.StatusOK, recorder.Code; expected != got {
			t.Errorf("expected code %v, got %v", expected, got)
		}
	})
}

func TestNewSheetFromTags(t *testing.T) {
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

func TestSheet_Validate(t *testing.T) {
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
