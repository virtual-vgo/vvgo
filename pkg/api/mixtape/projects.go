package mixtape

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	http2 "github.com/virtual-vgo/vvgo/pkg/api"
	"github.com/virtual-vgo/vvgo/pkg/api/errors"
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
		return nil, errors.InternalServerError
	}
	return saveProject(id, data, ctx)
}

func idFromUrl(url string) uint64 {
	s := strings.Split(url, "/")
	return redis.StringToObjectId(s[len(s)-1])
}

func ServeProjects(r *http.Request) http2.Response {
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		id := idFromUrl(r.URL.Path)
		projects, err := redis.ListMixtapeProjects(ctx)
		if err != nil {
			logger.MethodFailure(ctx, "models.ListMixtapeProjects", err)
			return errors.NewInternalServerError()
		}
		if id == 0 {
			return http2.Response{Status: http2.StatusOk, MixtapeProjects: projects}
		}
		for _, project := range projects {
			if project.Id == id {
				return http2.Response{Status: http2.StatusOk, MixtapeProject: &project}
			}
		}
		return errors.NewNotFoundError(fmt.Sprintf("id %d not found", id))

	case http.MethodPost:
		var data CreateProjectParams
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return errors.NewJsonDecodeError(err)
		}

		var id uint64
		if err := redis.Do(r.Context(), redis.Cmd(&id, redis.INCR, NextProjectIdRedisKey)); err != nil {
			logger.RedisFailure(ctx, err)
			return errors.NewInternalServerError()
		}
		return saveProject(id, data, ctx)

	case http.MethodPut:
		id := idFromUrl(r.URL.Path)
		if id == 0 {
			return errors.NewBadRequestError("invalid id")
		}

		var data CreateProjectParams
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return errors.NewJsonDecodeError(err)
		}
		return saveProject(id, data, ctx)

	case http.MethodDelete:
		id := idFromUrl(r.URL.Path)
		if id == 0 {
			return errors.NewBadRequestError("invalid id")
		}

		if err := redis.Do(ctx, redis.Cmd(nil, redis.HDEL, ProjectsRedisKey, redis.ObjectId(id).String())); err != nil {
			return errors.NewRedisError(err)
		}
		return http2.NewOkResponse()

	default:
		return errors.NewMethodNotAllowedError()
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
		return nil, errors.InternalServerError
	}

	if err := redis.Do(ctx, redis.Cmd(nil, redis.HSET,
		ProjectsRedisKey, redis.ObjectId(id).String(), projectJSON.String()),
	); err != nil {
		logger.RedisFailure(ctx, err)
		return nil, errors.RedisError(err)
	}

	return &project, nil
}
