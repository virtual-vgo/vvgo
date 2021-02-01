package api

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSlashCommand(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SlashCommand))
	req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(`{"type":1}`))
	require.NoError(t, err, "http.NewRequest() failed")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err, "http.Do()")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "status code")
	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	assert.Equal(t, `{"type":1}`, strings.TrimSpace(body.String()), "body")
}
