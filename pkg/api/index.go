package api

import (
	"github.com/virtual-vgo/vvgo/pkg/sheets"
	"net/http"
)

var IndexView = ServeTemplate("index.gohtml")
var ContactUs = ServeTemplate("contact_us.gohtml")

var AboutView = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	leaders, err := sheets.ListLeaders(ctx)
	if err != nil {
		logger.WithError(err).Error("readSheet() failed")
		internalServerError(w)
		return
	}
	ParseAndExecute(ctx, w, r, leaders, "about.gohtml")
})

func ServeTemplate(templateFile string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ParseAndExecute(r.Context(), w, r, nil, templateFile)
	}
}
