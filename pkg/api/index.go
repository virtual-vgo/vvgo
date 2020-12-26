package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

type IndexView struct{}

func (x IndexView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	ParseAndExecute(ctx, w, r, nil, "index.gohtml")
}

type AboutView struct{}

func (x AboutView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	leaders, err := sheets.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("readSheet() failed")
		internalServerError(w)
		return
	}
	ParseAndExecute(ctx, w, r, leaders, "about.gohtml")
}
