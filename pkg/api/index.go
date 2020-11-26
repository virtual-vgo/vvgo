package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets/leader"
	"net/http"
)

type IndexView struct{ Template }

func (x IndexView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	x.Template.ParseAndExecute(ctx, w, r, &struct{}{}, "index.gohtml")
}

type AboutView struct{ Template }

func (x AboutView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	leaders, err := leader.List(ctx, x.SpreadsheetID)
	if err != nil {
		logger.WithError(err).Error("readSheet() failed")
		internalServerError(w)
		return
	}
	x.Template.ParseAndExecute(ctx, w, r, leaders, "about.gohtml")
}
