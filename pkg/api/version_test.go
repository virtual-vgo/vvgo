package api

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/version", strings.NewReader(""))
	req.Header.Set("Accept", "application/json")
	Version(recorder, req)

	var gotJSON json.RawMessage
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"), "content type")
	assert.Equal(t, http.StatusOK, recorder.Code, "status code")
	assert.NoError(t, json.NewDecoder(recorder.Body).Decode(&gotJSON))
}
