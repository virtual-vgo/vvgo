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

	minioClient, err := vvgo_minio.NewClient(context.Background())
	require.NoError(t, err, "minio.New() failed")
	bucketName, err := minioClient.NewRandomBucket()
	require.NoError(t, err, "minioClient.MakeBucket() failed")
	_, err = minioClient.PutObject(bucketName, "danish", strings.NewReader(""), -1, minio.PutObjectOptions{})
	require.NoError(t, err, "minioClient.PutObject() failed")

	downloadHandler := DownloadHandler{}
	config := DownloadConfig{DistroBucket: bucketName}
	require.NoError(t, parse_config.WriteToRedisHash(ctx, "download", &config), "redis.Do() failed")

	for _, tt := range []struct {
		name    string
		request *http.Request
		wants   wants
	}{
		{
			name:    "post",
			request: httptest.NewRequest(http.MethodPost, "/download/danish", strings.NewReader("")),
			wants:   wants{code: http.StatusMethodNotAllowed},
		},
		{
			name:    "success",
			request: httptest.NewRequest(http.MethodGet, "/download/danish", strings.NewReader("")),
			wants:   wants{code: http.StatusFound},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			downloadHandler.ServeHTTP(recorder, tt.request)
			gotResp := recorder.Result()
			if expected, got := tt.wants.code, gotResp.StatusCode; expected != got {
				t.Errorf("expected code %v, got %v", expected, got)
			}
		})
	}
}
