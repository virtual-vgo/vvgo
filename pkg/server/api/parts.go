package api

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Parts(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)

	parts, err := models.ListParts(ctx, identity)
	if err != nil {
		logger.ListPartsFailure(ctx, err)
		return http_helpers.NewInternalServerError()
	}

	if parts == nil {
		parts = []models.Part{}
	}
	return models.ApiResponse{Status: models.StatusOk, Parts: parts.Sort()}
}
