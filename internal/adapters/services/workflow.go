package services

import (
	"github.com/google/uuid"
	"github.com/windevkay/flho/internal/domain"
)

type WorkflowService struct {
	repo domain.RunRepository
}

func NewWorkflowService(repo domain.RunRepository) *WorkflowService {
	return &WorkflowService{
		repo: repo,
	}
}

func (w *WorkflowService) StartWorkflowRun(workflowName string) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (w *WorkflowService) UpdateWorkflowRun(runID uuid.UUID, step int) error {
	return nil
}
