package test_helpers

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"net/http"
	"testing"
)

func AssertEqualApiResponses(t *testing.T, want, got models.ApiResponse) {
	t.Helper()

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	assert.NoError(t, encoder.Encode(got))
	gotJSON := buf.String()
	buf.Reset()
	assert.NoError(t, encoder.Encode(want))
	wantJSON := buf.String()
	assert.Equal(t, wantJSON, gotJSON)
}

func AssertEqualResponse(t *testing.T, want models.ApiResponse, got *http.Response) {
	t.Helper()
	require.NotNil(t, got, "got response is nil")
	if want.Status == models.StatusError {
		require.NotNil(t, want.Error, "error field")
		assert.Equal(t, want.Error.Code, got.StatusCode, "status code")
	} else {
		assert.Equal(t, http.StatusOK, got.StatusCode, "status code")
	}

	assert.Equal(t, "application/json", got.Header.Get("Content-Type"), "Content-Type")
	var gotResponse models.ApiResponse
	assert.NoError(t, json.NewDecoder(got.Body).Decode(&gotResponse), "json.Decode")
	var buf bytes.Buffer
	gotEncoder := json.NewEncoder(&buf)
	gotEncoder.SetIndent("", "  ")
	assert.NoError(t, gotEncoder.Encode(gotResponse))
	gotJSON := buf.String()
	buf.Reset()
	assert.NoError(t, gotEncoder.Encode(want))
	wantJSON := buf.String()
	assert.Equal(t, wantJSON, gotJSON)
}
