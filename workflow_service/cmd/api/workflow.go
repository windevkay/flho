package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/windevkay/flho/workflow_service/internal/data"
	"github.com/windevkay/flho/workflow_service/internal/services"
	errs "github.com/windevkay/flhoutils/errors"
	"github.com/windevkay/flhoutils/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO: Move this into flhoutils lib
func ReadObjectIDParam(r *http.Request) (primitive.ObjectID, error) {
	params := httprouter.ParamsFromContext(r.Context())
	idStr := params.ByName("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return primitive.NilObjectID, errors.New("invalid ID parameter")
	}

	return id, nil
}

func (app *application) createWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var input services.CreateWorkflowInput

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	ws := services.NewWorkflowService(workflowServiceConfig)
	workflow, err := ws.CreateWorkflow(input, app.contextGetUser(r))
	if err != nil {
		switch {
		case errors.Is(err.(*services.ValidationErr).Err, data.ErrValidationFailed):
			errs.FailedValidationResponse(w, r, err.(*services.ValidationErr).Fields)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	// handy resource location header for clients
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/workflows/%d", workflow.ID))

	helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"workflow": workflow}, headers)
}

func (app *application) showWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	id, err := ReadObjectIDParam(r)
	if err != nil {
		errs.NotFoundResponse(w, r)
		return
	}

	ws := services.NewWorkflowService(workflowServiceConfig)
	workflow, err := ws.ShowWorkflow(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			errs.NotFoundResponse(w, r)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"workflow": workflow}, nil)
}

func (app *application) updateWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	id, err := ReadObjectIDParam(r)
	if err != nil {
		errs.NotFoundResponse(w, r)
		return
	}

	var input services.UpdateInput

	err = helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	ws := services.NewWorkflowService(workflowServiceConfig)
	workflow, err := ws.UpdateWorkflow(id, input)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			errs.NotFoundResponse(w, r)
		case errors.Is(err.(*services.ValidationErr).Err, data.ErrValidationFailed):
			errs.FailedValidationResponse(w, r, err.(*services.ValidationErr).Fields)
		case errors.Is(err, data.ErrEditConflict):
			errs.EditConflictResponse(w, r)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"workflow": workflow}, nil)
}

func (app *application) deleteWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	id, err := ReadObjectIDParam(r)
	if err != nil {
		errs.NotFoundResponse(w, r)
		return
	}

	ws := services.NewWorkflowService(workflowServiceConfig)
	err = ws.DeleteWorkflow(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			errs.NotFoundResponse(w, r)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"message": "success: the workflow has been deleted"}, nil)
}

func (app *application) listWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var input services.ListWorkflowInput

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	ws := services.NewWorkflowService(workflowServiceConfig)
	workflows, metadata, err := ws.ListWorkflows(input, app.contextGetUser(r))
	if err != nil {
		switch {
		case errors.Is(err.(*services.ValidationErr).Err, data.ErrValidationFailed):
			errs.FailedValidationResponse(w, r, err.(*services.ValidationErr).Fields)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"workflows": workflows, "metadata": metadata}, nil)
}
