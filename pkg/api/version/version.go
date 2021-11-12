package version

import (
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
)

func Version(r *http.Request) api.Response {
	v := version.Get()
	switch r.Method {
	case http.MethodGet:
		return api.Response{Status: api.StatusOk, Version: &v}
	default:
		return response.NewMethodNotAllowedError()
	}
}
