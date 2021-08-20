package leaders

import (
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"net/http"
)

var logger = log.New()

func Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	leaders, err := models.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("sheets.ListLeaders() failed")
		helpers.InternalServerError(w)
		return
	}
	helpers.JsonEncode(w, &leaders)
}
