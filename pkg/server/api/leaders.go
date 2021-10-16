package api

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
)

func Leaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	leaders, err := models.ListLeaders(ctx)
	if err != nil {
		logger.ListLeadersFailure(ctx, err)
		http_helpers.InternalServerError(ctx, w)
		return
	}
	http_helpers.JsonEncode(w, &leaders)
}
