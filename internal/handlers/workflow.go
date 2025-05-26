package handlers

import "github.com/windevkay/flho/internal/domain"

type WorkflowHandler struct {
	service domain.WorkflowService
}

func NewWorkflowHandler(service domain.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{
		service: service,
	}
}
