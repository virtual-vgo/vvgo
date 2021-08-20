package leaders

import (
	"github.com/virtual-vgo/vvgo/pkg/api/helpers"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

var logger = log.New()

func Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	leaders, err := sheets.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("sheets.ListLeaders() failed")
		helpers.InternalServerError(w)
		return
	}
	helpers.JsonEncode(w, &leaders)
}
