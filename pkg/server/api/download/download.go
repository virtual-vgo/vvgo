package download

import (
	"github.com/virtual-vgo/vvgo/pkg/clients/minio"
	"github.com/virtual-vgo/vvgo/pkg/log"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"github.com/virtual-vgo/vvgo/pkg/server/helpers"
	"net/http"
	"time"
)

var logger = log.New()

const ProtectedLinkExpiry = 24 * 3600 * time.Second // 1 Day for protect links

func Handler(w http.ResponseWriter, r *http.Request) {
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
}
