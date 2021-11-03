package mixtape

import (
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"github.com/virtual-vgo/vvgo/pkg/server/login"
	"net/http"
)

type ProjectsPostRequest []models.MixtapeProject
type ProjectsDeleteRequest []string

func HandleProjects(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	identity := login.IdentityFromContext(ctx)
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
		var allowedProject []models.MixtapeProject
		for _, project := range projects {
			for _, owner := range project.Hosts {
				if identity.HasRole(models.RoleVVGOExecutiveDirector) || owner == identity.DiscordID {
					allowedProject = append(allowedProject, project)
					break
				}
			}

		}

		if err := models.WriteMixtapeProjects(ctx, allowedProject); err != nil {
			logger.MethodFailure(ctx, "models.WriteMixtapeProjects", err)
			return http_helpers.NewInternalServerError()
		}
		return models.ApiResponse{Status: models.StatusOk, MixtapeProjects: allowedProject}

	case http.MethodDelete:
		if !identity.HasRole(models.RoleVVGOExecutiveDirector) {
			return http_helpers.NewUnauthorizedError()
		}
		var args []string
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			return http_helpers.NewJsonDecodeError(err)
		}
		args = append([]string{models.MixtapeProjectsRedisKey}, args...)
		if err := redis.Do(ctx, redis.Cmd(nil, "HDEL", args...)); err != nil {
			logger.RedisFailure(ctx, err)
			return http_helpers.NewInternalServerError()
		}
		return models.ApiResponse{Status: models.StatusOk}

	default:
		return http_helpers.NewMethodNotAllowedError()
	}
}

