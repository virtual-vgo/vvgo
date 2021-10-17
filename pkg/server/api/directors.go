package api

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
)

func Directors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	directors, err := models.ListDirectors(ctx)
	if err != nil {
		logger.ListLeadersFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
		return
	}
	http_helpers.WriteAPIResponse(ctx, w, models.ApiResponse{
		Status:    models.StatusOk,
		Type:      models.ResponseTypeDirectors,
		Directors: &models.DirectorsResponse{Directors: directors},
	})
}
