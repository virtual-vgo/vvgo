package api

import (
	"github.com/virtual-vgo/vvgo/pkg/api/helpers"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

func LeadersApi(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	leaders, err := sheets.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("sheets.ListLeaders() failed")
		helpers.InternalServerError(w)
		return
	}
	helpers.JsonEncode(w, &leaders)
}
