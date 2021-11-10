package mixtape

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/models/mixtape"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
	"strings"
)

type CreateMixtapeProjectParams struct {
	Name    string   `json:"name"`
	Title   string   `json:"title"`
	Mixtape string   `json:"mixtape"`
	Blurb   string   `json:"blurb"`
	Channel string   `json:"channel"`
	Hosts   []string `json:"hosts,omitempty"`
}

type EditMixtapeProjectParams = CreateMixtapeProjectParams

func idFromUrl(url string) uint64 {
	s := strings.Split(url, "/")
	return redis.StringToObjectId(s[len(s)-1])
}

func HandleProjects(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		id := idFromUrl(r.URL.Path)
		projects, err := redis.ListMixtapeProjects(ctx)
		if err != nil {
			logger.MethodFailure(ctx, "models.ListMixtapeProjects", err)
			return http_helpers.NewInternalServerError()
		}
		if id == 0 {
			return models.ApiResponse{Status: models.StatusOk, MixtapeProjects: projects}
		}
		for _, project := range projects {
			if project.Id == id {
				return models.ApiResponse{Status: models.StatusOk, MixtapeProject: &project}
			}
		}
		return http_helpers.NewNotFoundError(fmt.Sprintf("id %s not found", id))

	case http.MethodPost:
		var data CreateMixtapeProjectParams
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return http_helpers.NewJsonDecodeError(err)
		}

		var id uint64
		if err := redis.Do(r.Context(), redis.Cmd(&id, redis.INCR, mixtape.NextProjectIdRedisKey)); err != nil {
			logger.RedisFailure(ctx, err)
			return http_helpers.NewInternalServerError()
		}
		return saveProject(id, data, ctx)

	case http.MethodPut:
		id := idFromUrl(r.URL.Path)
		if id == 0 {
			return http_helpers.NewBadRequestError("invalid id")
		}

		var data CreateMixtapeProjectParams
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return http_helpers.NewJsonDecodeError(err)
		}
		return saveProject(id, data, ctx)

	case http.MethodDelete:
		id := idFromUrl(r.URL.Path)
		if id == 0 {
			return http_helpers.NewBadRequestError("invalid id")
		}

		if err := redis.Do(ctx, redis.Cmd(nil, redis.HDEL, mixtape.ProjectsRedisKey, redis.ObjectId(id).String())); err != nil {
			return http_helpers.NewRedisError(err)
		}
		return http_helpers.NewOkResponse()

	default:
		return http_helpers.NewMethodNotAllowedError()
	}
}

func saveProject(id uint64, data CreateMixtapeProjectParams, ctx context.Context) models.ApiResponse {
	project := mixtape.Project{
		Id:      id,
		Name:    data.Name,
		Title:   data.Title,
		Mixtape: data.Mixtape,
		Blurb:   data.Blurb,
		Channel: data.Channel,
		Hosts:   data.Hosts,
	}
	var projectJSON bytes.Buffer
	if err := json.NewEncoder(&projectJSON).Encode(project); err != nil {
		logger.JsonEncodeFailure(ctx, err)
	}

	if err := redis.Do(ctx, redis.Cmd(nil, redis.HSET,
		mixtape.ProjectsRedisKey, redis.ObjectId(id).String(), projectJSON.String()),
	); err != nil {
		logger.RedisFailure(ctx, err)
		return http_helpers.NewRedisError(err)
	}

	return models.ApiResponse{Status: models.StatusOk, MixtapeProject: &project}
}
