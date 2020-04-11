package api

import (
	"bytes"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"github.com/virtual-vgo/vvgo/pkg/storage"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSheetsServer_ServeHTTP(t *testing.T) {
	type wants struct {
		code int
		body string
	}

	mockBucket := MockBucket{getObject: func(name string, dest *storage.Object) bool {
		if name == sheets.DataFile {
			*dest = storage.Object{
				ContentType: "application/json",
				Buffer:      *bytes.NewBuffer([]byte(`[]`)),
			}
		}
		return true
	}}

	for _, tt := range []struct {
		name    string
		request *http.Request
		wants   wants
	}{
		{
			name:    "method post",
			request: httptest.NewRequest(http.MethodPost, "/sheets", nil),
			wants:   wants{code: http.StatusMethodNotAllowed},
		},
		{
			name:    "method get",
			request: httptest.NewRequest(http.MethodGet, "/sheets", strings.NewReader("")),
			wants:   wants{code: http.StatusOK, body: `[]`},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := SheetsHandler{sheets.Sheets{Bucket: &mockBucket}}
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, tt.request)
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
