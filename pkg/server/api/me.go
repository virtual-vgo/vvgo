package api

import (
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Me(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)
	return models.ApiResponse{
		Status:   models.StatusOk,
		Identity: &identity,
	}
}
