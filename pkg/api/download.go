package api

import (
	"github.com/virtual-vgo/vvgo/pkg/minio"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"net/http"
	"time"
)

const ProtectedLinkExpiry = 24 * 3600 * time.Second // 1 Day for protect links

type DownloadHandler struct{}

type DownloadConfig struct {
	DistroBucket string `redis:"distro_bucket"`
}

func (x DownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	if err := parse_config.ReadFromRedisHash(ctx, "download", &config); err != nil {
		logger.WithError(err).Errorf("redis.Do() failed: %v", err)
		internalServerError(w)
		return
	}

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
}
