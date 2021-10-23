package mixtape

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
)

type PostRequest []models.MixtapeProject
type DeleteRequest []string

func Handler(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		projects, err := models.ListMixtapeProjects(ctx)
		if err != nil {
			logger.MethodFailure(ctx, "models.ListMixtapeProjects", err)
			return http_helpers.NewInternalServerError()
		}
		return models.ApiResponse{Status: models.StatusOk, MixtapeProjects: projects}

	case http.MethodPost:
		var projects []models.MixtapeProject
		if err := json.NewDecoder(r.Body).Decode(&projects); err != nil {
			return http_helpers.NewJsonDecodeError(err)
		}
		if err := models.WriteMixtapeProjects(ctx, projects); err != nil {
			logger.MethodFailure(ctx, "models.WriteMixtapeProjects", err)
			return http_helpers.NewInternalServerError()
		}
		return models.ApiResponse{Status: models.StatusOk}

	case http.MethodDelete:
		var args []string
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			return http_helpers.NewJsonDecodeError(err)
		}
		args = append([]string{"mixtape_projects"}, args...)
		if err := redis.Do(ctx, redis.Cmd(nil, "HDEL", args...)); err != nil {
			logger.RedisFailure(ctx, err)
			return http_helpers.NewInternalServerError()
		}
		return models.ApiResponse{Status: models.StatusOk}

	default:
		return http_helpers.NewMethodNotAllowedError()
	}
}
