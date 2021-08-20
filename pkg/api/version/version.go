package version

import (
	"github.com/virtual-vgo/vvgo/pkg/api/helpers"
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.MethodNotAllowed(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(version.JSON())
}
