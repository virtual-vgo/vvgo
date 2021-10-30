package api

import (
	"fmt"
	"github.com/minio/minio-go/v6"
	minio_wrapper "github.com/virtual-vgo/vvgo/pkg/clients/minio"
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

	fileName := r.URL.Query().Get("fileName")
	if fileName == "" {
		return http_helpers.NewBadRequestError("fileName is required")
	}

	minioClient, err := minio_wrapper.NewClient()
	if err != nil {
		logger.MethodFailure(ctx, "minio.New", err)
		return http_helpers.NewInternalServerError()
	}

	distroBucket := config.Config.VVGO.DistroBucket

	_, err = minioClient.StatObject(distroBucket, fileName, minio.StatObjectOptions{})
	if err != nil {
		logger.MethodFailure(ctx, "minio.StatObject", err)
		if _, ok := err.(minio.ErrorResponse); !ok {
			return http_helpers.NewInternalServerError()
		}
		switch err.(minio.ErrorResponse).StatusCode {
		case http.StatusNotFound:
			return http_helpers.NewNotFoundError(fmt.Sprintf("file `%s` not found", fileName))
		default:
			return http_helpers.NewInternalServerError()
		}
	}

	downloadUrl, err := minioClient.PresignedGetObject(distroBucket, fileName, ProtectedLinkExpiry, nil)
	if err != nil {
		logger.MethodFailure(ctx, "minio.StatObject", err)
		return http_helpers.NewInternalServerError()
	}
	return models.ApiResponse{Status: models.StatusFound, Location: downloadUrl.String()}
}
