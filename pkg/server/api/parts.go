package api

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

func Parts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)

	parts, err := models.ListParts(ctx, identity)
	if err != nil {
		logger.ListPartsFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
		return
	}

	if parts == nil {
		parts = []models.Part{}
	}
	parts = parts.Sort()
	http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
		Status: models.StatusOk,
		Parts:  parts,
	})
}
