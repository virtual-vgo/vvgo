package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

type LeadersAPI struct{}

func (x LeadersAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	leaders, err := sheets.ListLeaders(r.Context())
	if err != nil {
		logger.WithError(err).Error("readSheet() failed")
		internalServerError(w)
		return
	}
	jsonEncode(w, &leaders)
}
