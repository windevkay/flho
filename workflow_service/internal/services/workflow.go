package services

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"github.com/windevkay/flho/workflow_service/internal/data"
	"github.com/windevkay/flho/workflow_service/internal/rpc"
	"github.com/windevkay/flhoutils/helpers"
	"github.com/windevkay/flhoutils/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkflowService struct {
	config *WorkflowServiceConfig
}

type WorkflowServiceConfig struct {
	Models    data.Models
	Rpclients rpc.Clients
	Channel   *amqp091.Channel
	Wg        *sync.WaitGroup
	Logger    *slog.Logger
}

type CreateWorkflowInput struct {
	Name   string       `json:"name"`
	States []data.State `json:"states"`
}

type ListWorkflowInput struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type UpdateInput struct {
	Name *string `json:"name"`
}

type ValidationErr struct {
	Err    error
	Fields map[string]string
}

func (c *ValidationErr) Error() string { return "validation error" }

func NewWorkflowService(config *WorkflowServiceConfig) *WorkflowService {
	return &WorkflowService{
		config: config,
	}
}

func (ws *WorkflowService) CreateWorkflow(input CreateWorkflowInput, uuid string) (*data.Workflow, error) {
	v := validator.New()

	identityId, err := ws.config.Models.Identities.GetIdentityId(uuid)
	if err != nil || identityId == primitive.NilObjectID {
		ws.config.Logger.Error(err.Error())
		return nil, err
	}

	workflow := &data.Workflow{
		IdentityId: identityId,
		UniqueID:   helpers.GenerateUniqueId(15),
		Name:       input.Name,
		States:     input.States,
		Active:     true,
	}

	if data.ValidateWorkflow(v, workflow); !v.Valid() {
		ws.config.Logger.Error(fmt.Sprintf("validation failed: create workflow - %v", v.Errors))
		return nil, &ValidationErr{Err: data.ErrValidationFailed, Fields: v.Errors}
	}

	err = ws.config.Models.Workflows.Insert(workflow)
	if err != nil {
		ws.config.Logger.Error(err.Error())
		return nil, err
	}

	return workflow, nil
}

func (ws *WorkflowService) ShowWorkflow(id primitive.ObjectID) (*data.Workflow, error) {
	workflow, err := ws.config.Models.Workflows.Get(id)
	if err != nil {
		ws.config.Logger.Error(err.Error())
		return nil, err
	}

	return workflow, nil
}

func fullOrPartialUpdate(workflow *data.Workflow, input *UpdateInput) {
	if input.Name != nil {
		workflow.Name = *input.Name
	}
}

func (ws *WorkflowService) UpdateWorkflow(id primitive.ObjectID, input UpdateInput) (*data.Workflow, error) {
	workflow, err := ws.config.Models.Workflows.Get(id)
	if err != nil {
		ws.config.Logger.Error(err.Error())
		return nil, err
	}

	// achieve full or partial updates using non nil values
	fullOrPartialUpdate(workflow, &input)

	v := validator.New()

	if data.ValidateWorkflow(v, workflow); !v.Valid() {
		ws.config.Logger.Error(fmt.Sprintf("validation failed: update workflow - %v", v.Errors))
		return nil, &ValidationErr{Err: data.ErrValidationFailed, Fields: v.Errors}
	}

	err = ws.config.Models.Workflows.Update(workflow)
	if err != nil {
		ws.config.Logger.Error(err.Error())
		return nil, err
	}

	return workflow, nil
}

func (ws *WorkflowService) DeleteWorkflow(id primitive.ObjectID) error {
	return ws.config.Models.Workflows.Delete(id)
}

func (ws *WorkflowService) ListWorkflows(input ListWorkflowInput, uuid string) ([]*data.Workflow, *data.Metadata, error) {
	identityId, err := ws.config.Models.Identities.GetIdentityId(uuid)
	if err != nil || identityId == primitive.NilObjectID {
		ws.config.Logger.Error(err.Error())
		return nil, nil, err
	}

	v := validator.New()

	filter := data.Filters{
		Page:         input.Page,
		PageSize:     input.PageSize,
		Sort:         "-id",
		SortSafeList: []string{"id", "name", "-id", "-name"},
	}

	if data.ValidateFilters(v, filter); !v.Valid() {
		ws.config.Logger.Error(fmt.Sprintf("validation failed: list workflow - %v", v.Errors))
		return nil, nil, &ValidationErr{Err: data.ErrValidationFailed, Fields: v.Errors}
	}

	workflows, metadata, err := ws.config.Models.Workflows.GetAll(identityId, filter)
	return workflows, &metadata, err
}
