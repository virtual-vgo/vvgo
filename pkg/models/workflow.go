package models

import (
	"context"
)

type WorkflowStatus string

const (
	WorkflowStatusDone    = "done"
	WorkflowStatusSkipped = "skipped"
	WorkflowStatusFailed  = "failed"
)

type Workflow struct {
	Name  string
	Tasks []WorkflowTask
}

type WorkflowTask struct {
	Name string
	Do   func(ctx context.Context) error
}

type WorkflowTaskResult struct {
	Name    string
	Status  ApiResponseStatus
	Message string
}
