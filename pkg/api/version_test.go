package api

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	t.Run("accept:application/json", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/version", strings.NewReader(""))
		req.Header.Set("Accept", "application/json")
		Version(recorder, req)

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

		Version(recorder, req)
		if expected, got := http.StatusOK, recorder.Code; expected != got {
			t.Errorf("expected code %v, got %v", expected, got)
		}
		if expected, got := version.String(), recorder.Body.String(); expected != got {
			t.Errorf("expected code %v, got %v", expected, got)
		}
	})
}
