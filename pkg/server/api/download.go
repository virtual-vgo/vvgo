package api

import (
	"github.com/virtual-vgo/vvgo/pkg/clients/minio"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"time"
)

const ProtectedLinkExpiry = 24 * 3600 * time.Second // 1 Day for protect links

func Download(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		http_helpers.WriteErrorMethodNotAllowed(ctx, w)
		return
	}

	object := r.URL.Query().Get("object")
	if object == "" {
		http_helpers.WriteErrorBadRequest(ctx, w, "object required")
		return
	}

	minioClient, err := minio.NewClient()
	if err != nil {
		logger.MethodFailure(ctx, "minio.New", err)
		http_helpers.WriteInternalServerError(ctx, w)
		return
	}

	distroBucket := config.Config.VVGO.DistroBucket
	downloadUrl, err := minioClient.PresignedGetObject(distroBucket, object, ProtectedLinkExpiry, nil)
	if err != nil {
		logger.MethodFailure(ctx, "minio.StatObject", err)
		http_helpers.WriteInternalServerError(ctx, w)
		return
	}
	http.Redirect(w, r, downloadUrl.String(), http.StatusFound)
}
