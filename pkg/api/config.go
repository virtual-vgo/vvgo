package api

import (
	"io"
	"net/http"
	"os"
)

const ConfigFile = "/etc/vvgo/vvgo.json"

var ConfigApi = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	file, err := os.Open(ConfigFile)
	if err != nil {
		logger.SomeMethodFailure(r.Context(), "os.Open", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err := io.Copy(w, file); err != nil {
		logger.SomeMethodFailure(r.Context(), "io.Copy", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
})
