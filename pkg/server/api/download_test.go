package api

import (
	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vvgo_minio "github.com/virtual-vgo/vvgo/pkg/clients/minio"
	"github.com/virtual-vgo/vvgo/pkg/config"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers/test_helpers"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDownload(t *testing.T) {
	minioClient, err := vvgo_minio.NewClient()
	require.NoError(t, err, "minio.New() failed")
	bucketName, err := minioClient.NewRandomBucket()
	require.NoError(t, err, "minioClient.MakeBucket() failed")
	_, err = minioClient.PutObject(bucketName, "danish", strings.NewReader(""), -1, minio.PutObjectOptions{})
	require.NoError(t, err, "minioClient.PutObject() failed")
	config.Config.VVGO.DistroBucket = bucketName

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/download?fileName=danish", nil)
		test_helpers.AssertEqualApiResponses(t, http_helpers.NewMethodNotAllowedError(), Download(req))
	})

	t.Run("fileName is empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/download", nil)
		test_helpers.AssertEqualApiResponses(t, http_helpers.NewBadRequestError("fileName is required"), Download(req))
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/download?fileName=danishxx", nil)
		test_helpers.AssertEqualApiResponses(t, http_helpers.NewNotFoundError("file `danishxx` not found"), Download(req))
	})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/download?fileName=danish", nil)
		resp := Download(req)
		assert.NotEmpty(t, resp.Location, "location")
		assert.Equal(t, models.StatusFound, resp.Status, "status")
		assert.Nil(t, resp.Error)
	})
}
