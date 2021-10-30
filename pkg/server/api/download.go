package api

import (
	"github.com/virtual-vgo/vvgo/pkg/clients/minio"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"time"
)

const ProtectedLinkExpiry = 24 * 3600 * time.Second // 1 Day for protect links

type GetDownloadRequest struct {
	FileName string
}

func Download(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		return http_helpers.NewMethodNotAllowedError()
	}

	object := r.URL.Query().Get("fileName")
	if object == "" {
		return http_helpers.NewBadRequestError("fileName is required")
	}

	minioClient, err := minio.NewClient()
	if err != nil {
		logger.MethodFailure(ctx, "minio.New", err)
		return http_helpers.NewInternalServerError()
	}

	distroBucket := config.Config.VVGO.DistroBucket
	downloadUrl, err := minioClient.PresignedGetObject(distroBucket, object, ProtectedLinkExpiry, nil)
	if err != nil {
		logger.MethodFailure(ctx, "minio.StatObject", err)
		return http_helpers.NewInternalServerError()
	}
	return models.ApiResponse{Status: models.StatusFound, Location: downloadUrl.String()}
}
