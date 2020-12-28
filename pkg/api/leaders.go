package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

type Leaders struct{}

func (x Leaders) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	leaders, err := sheets.ListLeaders(r.Context())
	if err != nil {
		logger.WithError(err).Error("readSheet() failed")
		internalServerError(w)
		return
	}
	jsonEncode(w, &leaders)
}
