package version

import (
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/test_helpers"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	got := Version(req)

	wantVersion := version.Get()
	test_helpers.AssertEqualApiResponses(t, api.Response{
		Status:  api.StatusOk,
		Version: &wantVersion,
	}, got)
}
