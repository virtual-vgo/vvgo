package download

import (
	"fmt"
	"github.com/minio/minio-go/v6"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	minio_wrapper "github.com/virtual-vgo/vvgo/pkg/clients/minio"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"time"
)

const ProtectedLinkExpiry = 24 * 3600 * time.Second // 1 Day for protect links

type GetDownloadRequest struct {
	FileName string
}

func Download(r *http.Request) api.Response {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		return response.NewMethodNotAllowedError()
	}

	fileName := r.URL.Query().Get("fileName")
	if fileName == "" {
		return response.NewBadRequestError("fileName is required")
	}

	minioClient, err := minio_wrapper.NewClient()
	if err != nil {
		logger.MethodFailure(ctx, "minio.New", err)
		return response.NewInternalServerError()
	}

	distroBucket := config.Env.VVGO.DistroBucket

	_, err = minioClient.StatObject(distroBucket, fileName, minio.StatObjectOptions{})
	if err != nil {
		logger.MethodFailure(ctx, "minio.StatObject", err)
		if _, ok := err.(minio.ErrorResponse); !ok {
			return response.NewInternalServerError()
		}
		switch err.(minio.ErrorResponse).StatusCode {
		case http.StatusNotFound:
			return response.NewNotFoundError(fmt.Sprintf("file `%s` not found", fileName))
		default:
			return response.NewInternalServerError()
		}
	}

	downloadUrl, err := minioClient.PresignedGetObject(distroBucket, fileName, ProtectedLinkExpiry, nil)
	if err != nil {
		logger.MethodFailure(ctx, "minio.StatObject", err)
		return response.NewInternalServerError()
	}
	return api.Response{Status: api.StatusFound, Location: downloadUrl.String()}
}
