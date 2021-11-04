package mixtape

import (
	"context"
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
		if project.Name == want {
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
		var status models.ApiResponseStatus
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
