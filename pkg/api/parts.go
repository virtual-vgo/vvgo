package api

import (
	"net/http"
)

type PartView struct{}

func (x PartView) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	ParseAndExecute(ctx, w, r, nil, "parts.gohtml")
}
