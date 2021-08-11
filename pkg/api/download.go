package api

import (
	"github.com/virtual-vgo/vvgo/pkg/minio"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"net/http"
	"time"
)

const ProtectedLinkExpiry = 24 * 3600 * time.Second // 1 Day for protect links

type DownloadConfig struct {
	DistroBucket string `json:"distro_bucket" default:"vvgo-distro"`
}

var DownloadHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	object := r.URL.Query().Get("object")
	if object == "" {
		badRequest(w, "object required")
		return
	}

	ctx := r.Context()
	var config DownloadConfig
	parse_config.ReadConfigModule(ctx, "download", &config)

	minioClient, err := minio.NewClient(ctx)
	if err != nil {
		logger.WithError(err).Error("minio.New() failed")
		internalServerError(w)
		return
	}

	downloadUrl, err := minioClient.PresignedGetObject(config.DistroBucket, object, ProtectedLinkExpiry, nil)
	if err != nil {
		logger.WithError(err).Error("minio.StatObject() failed")
		internalServerError(w)
		return
	}
	http.Redirect(w, r, downloadUrl.String(), http.StatusFound)
})
