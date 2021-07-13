package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

func LeadersApi(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	leaders, err := sheets.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("sheets.ListLeaders() failed")
		internalServerError(w)
		return
	}
	jsonEncode(w, &leaders)
}

func AboutMeApi(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	entries, err := readAboutMeEntries(ctx)
	if err != nil {
		logger.WithError(err).Error("readAboutMeEntries() failed")
		internalServerError(w)
		return
	}
	jsonEncode(w, &entries)
}
