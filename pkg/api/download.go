package api

import (
	"github.com/minio/minio-go/v6"
	"net/http"
)

func (x *Server) Download(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	values := r.URL.Query()
	key := values.Get("key")
	bucket := values.Get("bucket")

	downloadURL, err := x.DownloadURL(bucket, key)
	switch e := err.(type) {
	case nil:
		http.Redirect(w, r, downloadURL, http.StatusFound)
	case minio.ErrorResponse:
		if e.StatusCode == http.StatusNotFound {
			http.NotFound(w, r)
		} else {
			http.Error(w, "", http.StatusInternalServerError)
		}
	default:
		logger.WithError(err).Error("minio.StatObject() failed")
		http.Error(w, "", http.StatusInternalServerError)
	}
}
