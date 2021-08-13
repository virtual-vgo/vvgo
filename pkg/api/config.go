package api

import (
	"io"
	"net/http"
	"os"
)

const ConfigFile = "/etc/vvgo/vvgo.json"

var ConfigApi = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open(ConfigFile)
	if err != nil {
		logger.Errorf("os.Open() failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err := io.Copy(w, file); err != nil {
		logger.Errorf("io.Copy() failed: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
})
