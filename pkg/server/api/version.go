package api

import (
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
)

func Version(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.MethodNotAllowed(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(version.JSON())
}
