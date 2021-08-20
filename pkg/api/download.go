package api

import (
	"github.com/virtual-vgo/vvgo/pkg/api/helpers"
	"github.com/virtual-vgo/vvgo/pkg/minio"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"net/http"
	"time"
)

const ProtectedLinkExpiry = 24 * 3600 * time.Second // 1 Day for protect links

type DownloadConfig struct {
	DistroBucket string `json:"distro_bucket" envconfig:"distro_bucket" default:"vvgo-distro"`
}

var DownloadHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		helpers.MethodNotAllowed(w)
		return
	}

	object := r.URL.Query().Get("object")
	if object == "" {
		helpers.BadRequest(w, "object required")
		return
	}

	minioClient, err := minio.NewClient()
	if err != nil {
		logger.WithError(err).Error("minio.New() failed")
		helpers.InternalServerError(w)
		return
	}

	distroBucket := parse_config.Config.VVGO.DistroBucket
	downloadUrl, err := minioClient.PresignedGetObject(distroBucket, object, ProtectedLinkExpiry, nil)
	if err != nil {
		logger.WithError(err).Error("minio.StatObject() failed")
		helpers.InternalServerError(w)
		return
	}
	http.Redirect(w, r, downloadUrl.String(), http.StatusFound)
})
