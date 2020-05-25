package api

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virtual-vgo/vvgo/pkg/projects"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProjectsHandler_ServeHTTP(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		ts := httptest.NewServer(ProjectsHandler{})
		defer ts.Close()
		resp, err := http.Get(ts.URL + "?name=01-snake-eater")
		require.NoError(t, err, "http.Get")
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var got projects.Project
		assert.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		assert.Equal(t, &got, projects.GetName("01-snake-eater"))
	})

	t.Run("not found", func(t *testing.T) {
		ts := httptest.NewServer(ProjectsHandler{})
		defer ts.Close()
		resp, err := http.Get(ts.URL + "?name=00-might-morphin-power-rangers")
		require.NoError(t, err, "http.Get")
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
