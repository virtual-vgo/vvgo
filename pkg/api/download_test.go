package api

import (
	"context"
	"github.com/minio/minio-go/v6"
	"github.com/stretchr/testify/require"
	vvgo_minio "github.com/virtual-vgo/vvgo/pkg/minio"
	"github.com/virtual-vgo/vvgo/pkg/parse_config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDownloadHandler_ServeHTTP(t *testing.T) {
	ctx := context.Background()
	type wants struct{ code int }

	minioClient, err := vvgo_minio.NewClient()
	require.NoError(t, err, "minio.New() failed")
	bucketName, err := minioClient.NewRandomBucket()
	require.NoError(t, err, "minioClient.MakeBucket() failed")
	_, err = minioClient.PutObject(bucketName, "danish", strings.NewReader(""), -1, minio.PutObjectOptions{})
	require.NoError(t, err, "minioClient.PutObject() failed")

	downloadHandler := DownloadHandler
	parse_config.Config.VVGO.DistroBucket=bucketName

	for _, tt := range []struct {
		name    string
		request *http.Request
		wants   wants
	}{
		{
			name:    "post",
			request: httptest.NewRequest(http.MethodPost, "/download?object=danish", strings.NewReader("")),
			wants:   wants{code: http.StatusMethodNotAllowed},
		},
		{
			name:    "success",
			request: httptest.NewRequest(http.MethodGet, "/download?object=danish", strings.NewReader("")),
			wants:   wants{code: http.StatusFound},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			downloadHandler.ServeHTTP(recorder, tt.request.WithContext(ctx))
			gotResp := recorder.Result()
			if expected, got := tt.wants.code, gotResp.StatusCode; expected != got {
				t.Errorf("expected code %v, got %v", expected, got)
			}
		})
	}
}
