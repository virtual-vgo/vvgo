package api

import (
	"github.com/virtual-vgo/vvgo/pkg/version"
	"net/http"
)

func (x Server) Version(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	versionHeader := version.Header()
	for k := range versionHeader {
		w.Header().Set(k, versionHeader.Get(k))
	}

	switch true {
	case acceptsType(r, "application/json"):
		w.Write(version.JSON())
	default:
		w.Write([]byte(version.String()))
	}
}
