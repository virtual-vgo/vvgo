package api

import (
	"context"
	"github.com/minio/minio-go/v6"
	"github.com/virtual-vgo/vvgo/pkg/tracing"
	"net/http"
)

// Accepts query params `object` and `bucket`.
// The map key is the bucket param.
// The map value function should return the url of the object and any error encountered.
type DownloadHandler map[string]func(ctx context.Context, objectName string) (url string, err error)

func (x DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracing.StartSpan(r.Context(), "download_handler")
	defer span.Send()

	if r.Method != http.MethodGet {
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}

	values := r.URL.Query()
	objectName := values.Get("object")
	bucketName := values.Get("bucket")

	if bucketName == "" {
		badRequest(w, "bucket required")
	}
	if objectName == "" {
		badRequest(w, "object required")
	}

	urlFunc, ok := x[bucketName]
	if !ok {
		unauthorized(w)
		return
	}

	downloadURL, err := urlFunc(ctx, objectName)
	switch e := err.(type) {
	case nil:
		http.Redirect(w, r, downloadURL, http.StatusFound)
	case minio.ErrorResponse:
		if e.StatusCode == http.StatusNotFound {
			notFound(w)
		} else {
			internalServerError(w)
		}
	default:
		logger.WithError(err).Error("minio.StatObject() failed")
		http.Error(w, "", http.StatusInternalServerError)
	}
}
