package api

import (
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/minio"
	"net/http"
	"time"
)

const ProtectedLinkExpiry = 24 * 3600 * time.Second // 1 Day for protect links

type DownloadHandler struct{}

func (x DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	values := r.URL.Query()
	object := values.Get("object")
	if object == "" {
		badRequest(w, "object required")
		return
	}

	ctx := r.Context()
	bucket := parse_config.DistroBucket(ctx)
	minioClient, err := minio.NewClient(ctx)
	if err != nil {
		logger.WithError(err).Error("minio.New() failed")
		internalServerError(w)
		return
	}

	downloadUrl, err := minioClient.PresignedGetObject(bucket, object, ProtectedLinkExpiry, nil)
	if err != nil {
		logger.WithError(err).Error("minio.StatObject() failed")
		internalServerError(w)
		return
	}
	http.Redirect(w, r, downloadUrl.String(), http.StatusFound)
}
