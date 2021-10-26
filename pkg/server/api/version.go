package api

import (
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
)

func Version(r *http.Request) models.ApiResponse {
	switch r.Method {
	case http.MethodGet:
		return models.ApiResponse{Status: models.StatusOk, Version: version.Get()}
	default:
		return http_helpers.NewMethodNotAllowedError()
	}
}
