package api

import (
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"net/http"
)

func Leaders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	leaders, err := models.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("sheets.ListLeaders() failed")
		helpers.InternalServerError(w)
		return
	}
	helpers.JsonEncode(w, &leaders)
}
