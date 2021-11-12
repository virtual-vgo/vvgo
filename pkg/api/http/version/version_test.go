package version

import (
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/test_helpers"
	"github.com/virtual-vgo/vvgo/pkg/api/version"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersion(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	got := Version(req)

	wantVersion := version.Get()
	test_helpers.AssertEqualApiResponses(t, http2.Response{
		Status:  http2.StatusOk,
		Version: &wantVersion,
	}, got)
}
