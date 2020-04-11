package api

import (
	"github.com/minio/minio-go/v6"
	"net/http"
)

// Accepts query params `object` and `bucket`.
// The map key is the bucket param.
// The map value function should return the url of the object and any error encountered.
type DownloadHandler map[string]func(objectName string) (url string, err error)

func (x DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	values := r.URL.Query()
	objectName := values.Get("object")
	bucketName := values.Get("bucket")

	urlFunc, ok := x[bucketName]
	if !ok {
		unauthorized(w, r)
		return
	}

	downloadURL, err := urlFunc(objectName)
	switch e := err.(type) {
	case nil:
		http.Redirect(w, r, downloadURL, http.StatusFound)
	case minio.ErrorResponse:
		if e.StatusCode == http.StatusNotFound {
			notFound(w, r)
		} else {
			internalServerError(w)
		}
	default:
		logger.WithError(err).Error("minio.StatObject() failed")
		http.Error(w, "", http.StatusInternalServerError)
	}
}
