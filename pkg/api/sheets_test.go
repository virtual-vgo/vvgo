package api

import (
	"bytes"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSheetsServer_ServeHTTP(t *testing.T) {
	type request struct {
		method  string
		body    string
		accepts string
	}
	type wants struct {
		code int
		body string
	}
	mockBodyBytes, err := ioutil.ReadFile("testdata/sheets.html")
	if err != nil {
		t.Fatalf("ioutil.ReadFile() failed: %v", err)
	}
	mockHTML := string(mockBodyBytes)
	mockJSON := `[{"project":"truly",
					"part_name":"dio-brando",
					"part_number":3,
					"file_key":"0xff",
					"link":"/download?bucket=sheets\u0026object=0xff"}]`

	mockBucket := MockBucket{getObject: func(name string, dest *storage.Object) bool {
		if name == sheets.DataFile {
			sheets := []sheets.Sheet{{
				Project:    "truly",
				PartName:   "dio-brando",
				PartNumber: 3,
				FileKey:    "0xff",
			}}
			var buffer bytes.Buffer
			json.NewEncoder(&buffer).Encode(sheets)
			*dest = storage.Object{
				ContentType: "application/json",
				Buffer:      buffer,
			}
		}
		return true
	}}

	for _, tt := range []struct {
		name    string
		request request
		wants   wants
	}{
		{
			name: "method post",
			request: request{
				method: http.MethodPost,
			},
			wants: wants{
				code: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "method get",
			request: request{
				method: http.MethodGet,
			},
			wants: wants{
				code: http.StatusOK,
				body: mockJSON,
			},
		},
		{
			name: "method get/accept text/html",
			request: request{
				method:  http.MethodGet,
				accepts: "text/html",
				body:    mockHTML,
			},
			wants: wants{
				code: http.StatusOK,
				body: mockHTML,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := SheetsHandler{sheets.Sheets{Bucket: &mockBucket}}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(tt.request.method, "/sheets", strings.NewReader(tt.request.body))
			request.Header.Set("Accept", tt.request.accepts)
			server.ServeHTTP(recorder, request)
			gotResp := recorder.Result()
			gotBody := strings.TrimSpace(recorder.Body.String())
			if expected, got := tt.wants.code, gotResp.StatusCode; expected != got {
				t.Errorf("expected code %v, got %v", expected, got)
			}

			switch tt.wants.body {
			case mockJSON:
				var buf []map[string]interface{}
				json.Unmarshal([]byte(mockJSON), &buf)
				wantBytes, _ := json.Marshal(&buf)
				tt.wants.body = string(wantBytes)
				json.Unmarshal([]byte(gotBody), &buf)
				gotBytes, _ := json.Marshal(&buf)
				gotBody = string(gotBytes)

			case mockHTML:
				wantHTML := html.NewTokenizer(strings.NewReader(mockHTML))
				gotHTML := html.NewTokenizer(strings.NewReader(gotBody))
				tt.wants.body = ""
				for token := wantHTML.Next(); token != html.ErrorToken; token = wantHTML.Next() {
					tt.wants.body += string(bytes.TrimSpace(wantHTML.Raw()))
				}

				gotBody = ""
				for token := gotHTML.Next(); token != html.ErrorToken; token = gotHTML.Next() {
					gotBody += string(bytes.TrimSpace(gotHTML.Raw()))
				}
			}

			if expected, got := tt.wants.body, gotBody; expected != got {
				t.Errorf("expected body:\nwant: `%s`\n got: `%s`", expected, got)
			}
		})
	}
}
