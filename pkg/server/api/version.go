package api

import (
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
)

func Version(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		http_helpers.WriteErrorMethodNotAllowed(ctx, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(version.JSON())
}
