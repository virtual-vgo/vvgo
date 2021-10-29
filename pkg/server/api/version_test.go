package api

import (
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers/test_helpers"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	got := Version(req)

	wantVersion := version.Get()
	test_helpers.AssertEqualApiResponses(t, models.ApiResponse{
		Status:  models.StatusOk,
		Version: &wantVersion,
	}, got)
}
