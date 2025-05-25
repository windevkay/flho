package domain

import "github.com/google/uuid"

type Timeout int

type State struct {
	Name       string
	Step       int
	Retry      bool
	RetryUrl   string
	RetryAfter Timeout
}

type Workflow struct {
	ID     uuid.UUID
	Name   string
	States []State
}

type Run struct {
	UniqueID uuid.UUID
	Step     int
}

type WorkflowService interface {
	// StartWorkflowRun starts a run of the workflow and returns the run unique id.
	// It is flexible and can start a run from a step other than 0.
	StartWorkflowRun(name string, step int) (uuid.UUID, error)
	// UpdateWorkflowRun updates the run step to reflect the current state of the workflow
	UpdateWorkflowRun(runID uuid.UUID, step int) error
	ListCurrentRunsForWorkflow(workflowID uuid.UUID) ([]Run, error)
}

type RunRepository interface {
	SaveRun(run *Run) error
	UpdateRun(runID uuid.UUID, step int) error
}
