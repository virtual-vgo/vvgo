package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

func LeadersApi(w http.ResponseWriter, r *http.Request) {
	leaders, err := sheets.ListLeaders(r.Context())
	if err != nil {
		logger.WithError(err).Error("sheets.ListLeaders() failed")
		internalServerError(w)
	}
	jsonEncode(w, &leaders)
}
