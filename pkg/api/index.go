package api

import (
	"net/http"
)

var IndexView = ServeTemplate("index.gohtml")
var ContactUs = ServeTemplate("contact_us.gohtml")
var AboutView = ServeTemplate("about.gohtml")
var PartsView = ServeTemplate("parts.gohtml")
var VotingView = ServeTemplate("voting.gohtml")

func ServeTemplate(templateFile string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ParseAndExecute(r.Context(), w, r, nil, templateFile)
	}
}
