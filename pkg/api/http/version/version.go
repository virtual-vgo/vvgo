package version

import (
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
	"github.com/virtual-vgo/vvgo/pkg/api/version"
	"net/http"
)

func Version(r *http.Request) http2.Response {
	v := version.Get()
	switch r.Method {
	case http.MethodGet:
		return http2.Response{Status: http2.StatusOk, Version: &v}
	default:
		return errors.NewMethodNotAllowedError()
	}
}
