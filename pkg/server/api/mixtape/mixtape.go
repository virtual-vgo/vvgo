package mixtape

import (
	"context"
	"encoding/json"
	"github.com/virtual-vgo/vvgo/pkg/clients/redis"
	"github.com/virtual-vgo/vvgo/pkg/errors"
	"github.com/virtual-vgo/vvgo/pkg/logger"
	"github.com/virtual-vgo/vvgo/pkg/models"
	"github.com/virtual-vgo/vvgo/pkg/server/http_helpers"
	"net/http"
)

func buildCreateNewMixtapeProjectWorkflow(project models.MixtapeProject) models.Workflow {
	return models.Workflow{
		Name: "Create New Mixtape Project",
		Tasks: []models.WorkflowTask{
			{
				Name: "Create project channel in Discord",
				Do:   func(ctx context.Context) error { return errors.New("implement me") },
			},
			{
				Name: "Give project owners role MixtapeHost",
				Do:   func(ctx context.Context) error { return errors.New("implement me") },
			},
			{
				Name: "Add project owners to the project channel",
				Do:   func(ctx context.Context) error { return errors.New("implement me") },
			},
		},
	}
}

func WorkflowHandler(r *http.Request) models.ApiResponse {
	ctx := r.Context()
	want := r.URL.Query().Get("projectId")
	if want == "" {
		return http_helpers.NewBadRequestError("projectId cannot be empty")
	}
	projects, err := models.ListMixtapeProjects(ctx)
	if err != nil {
		logger.MethodFailure(ctx, "models.ListMixtapeProjects", err)
		return http_helpers.NewInternalServerError()
	}
	var wantProject *models.MixtapeProject
	for _, project := range projects {
		if project.Id == want {
			wantProject = &project
			break
		}
	}
	if wantProject == nil {
		return http_helpers.NewBadRequestError("project not found")
	}
	workflow := buildCreateNewMixtapeProjectWorkflow(*wantProject)
	var results []models.WorkflowTaskResult
	for _, task := range workflow.Tasks {
		var status string
		var message string
		if err := task.Do(ctx); err != nil {
			message = err.Error()
			status = models.WorkflowStatusFailed
		} else {
			status = models.StatusOk
		}

		results = append(results, models.WorkflowTaskResult{
			Name:    task.Name,
			Status:  status,
			Message: message,
		})
	}
	return models.ApiResponse{Status: models.StatusOk, WorkflowResult: results}
}

type ProjectsPostRequest []models.MixtapeProject
type ProjectsDeleteRequest []string

func ProjectsHandler(r *http.Request) models.ApiResponse {
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
		args = append([]string{models.MixtapeRedisKey}, args...)
		if err := redis.Do(ctx, redis.Cmd(nil, "HDEL", args...)); err != nil {
			logger.RedisFailure(ctx, err)
			return http_helpers.NewInternalServerError()
		}
		return models.ApiResponse{Status: models.StatusOk}

	default:
		return http_helpers.NewMethodNotAllowedError()
	}
}
