package mixtape

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/response"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"net/http"
	"strings"
)

const ProjectsRedisKey = "mixtape:projects"
const NextProjectIdRedisKey = "mixtape:next_project_id"

type Project struct {
	Id      uint64   `json:"id"`
	Name    string   `json:"Name"`
	Title   string   `json:"title"`
	Mixtape string   `json:"mixtape"`
	Blurb   string   `json:"blurb"`
	Channel string   `json:"channel"`
	Hosts   []string `json:"hosts,omitempty"`
}

type CreateProjectParams struct {
	Name    string   `json:"name"`
	Title   string   `json:"title"`
	Mixtape string   `json:"mixtape"`
	Blurb   string   `json:"blurb"`
	Channel string   `json:"channel"`
	Hosts   []string `json:"hosts,omitempty"`
}

func CreateProject(ctx context.Context, data CreateProjectParams) (*Project, error) {
	var id uint64
	if err := redis.Do(ctx, redis.Cmd(&id, redis.INCR, NextProjectIdRedisKey)); err != nil {
		logger.RedisFailure(ctx, err)
		return nil, response.InternalServerError
	}
	return saveProject(id, data, ctx)
}

func idFromUrl(url string) uint64 {
	s := strings.Split(url, "/")
	return redis.StringToObjectId(s[len(s)-1])
}

func ServeProjects(r *http.Request) api.Response {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		id := idFromUrl(r.URL.Path)
		projects, err := redis.ListMixtapeProjects(ctx)
		if err != nil {
			logger.MethodFailure(ctx, "models.ListMixtapeProjects", err)
			return response.NewInternalServerError()
		}
		if id == 0 {
			return api.Response{Status: api.StatusOk, MixtapeProjects: projects}
		}
		for _, project := range projects {
			if project.Id == id {
				return api.Response{Status: api.StatusOk, MixtapeProject: &project}
			}
		}
		return response.NewNotFoundError(fmt.Sprintf("id %d not found", id))

	case http.MethodPost:
		var data CreateProjectParams
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return response.NewJsonDecodeError(err)
		}

		var id uint64
		if err := redis.Do(r.Context(), redis.Cmd(&id, redis.INCR, NextProjectIdRedisKey)); err != nil {
			logger.RedisFailure(ctx, err)
			return response.NewInternalServerError()
		}
		return saveProject(id, data, ctx)

	case http.MethodPut:
		id := idFromUrl(r.URL.Path)
		if id == 0 {
			return response.NewBadRequestError("invalid id")
		}

		var data CreateProjectParams
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return response.NewJsonDecodeError(err)
		}
		return saveProject(id, data, ctx)

	case http.MethodDelete:
		id := idFromUrl(r.URL.Path)
		if id == 0 {
			return response.NewBadRequestError("invalid id")
		}

		if err := redis.Do(ctx, redis.Cmd(nil, redis.HDEL, ProjectsRedisKey, redis.ObjectId(id).String())); err != nil {
			return response.NewRedisError(err)
		}
		return api.NewOkResponse()

	default:
		return response.NewMethodNotAllowedError()
	}
}

func saveProject(id uint64, data CreateProjectParams, ctx context.Context) (*Project, error) {
	project := Project{
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
		return nil, response.InternalServerError
	}

	if err := redis.Do(ctx, redis.Cmd(nil, redis.HSET,
		ProjectsRedisKey, redis.ObjectId(id).String(), projectJSON.String()),
	); err != nil {
		logger.RedisFailure(ctx, err)
		return nil, response.RedisError(err)
	}

	return &project, nil
}
