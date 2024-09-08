package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/windevkay/flho/workflow_service/internal/data"
	errs "github.com/windevkay/flhoutils/errors"
	"github.com/windevkay/flhoutils/helpers"
	"github.com/windevkay/flhoutils/validator"
)

func (app *application) createWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name   string       `json:"name"`
		States []data.State `json:"states"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	identityId, err := app.models.Workflows.GetIdentityId(app.contextGetUser(r))
	if err != nil || identityId <= 0 {
		errs.ServerErrorResponse(w, r, err)
		return
	}

	workflow := &data.Workflow{
		IdentityId: identityId,
		UniqueID:   helpers.GenerateUniqueId(15),
		Name:       input.Name,
		States:     input.States,
		Active:     true,
	}

	if data.ValidateWorkflow(v, workflow); !v.Valid() {
		errs.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Workflows.InsertWithTx(workflow)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}

	// handy resource location header for clients
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/workflows/%d", workflow.ID))

	helpers.WriteJSON(w, http.StatusCreated, helpers.Envelope{"workflow": workflow}, headers)
}

func (app *application) showWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.ReadIDParam(r)
	if err != nil {
		errs.NotFoundResponse(w, r)
		return
	}

	workflow, err := app.models.Workflows.Get(id)
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

func fullOrPartialUpdate(workflow *data.Workflow, input *struct {
	Name *string `json:"name"`
}) {
	if input.Name != nil {
		workflow.Name = *input.Name
	}
}

func (app *application) updateWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	id, err := helpers.ReadIDParam(r)
	if err != nil {
		errs.NotFoundResponse(w, r)
		return
	}

	workflow, err := app.models.Workflows.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			errs.NotFoundResponse(w, r)
		default:
			errs.ServerErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name *string `json:"name"`
	}

	err = helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	// achieve full or partial updates using non nil values
	fullOrPartialUpdate(workflow, &input)

	v := validator.New()

	if data.ValidateWorkflow(v, workflow); !v.Valid() {
		errs.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Workflows.Update(workflow)
	if err != nil {
		switch {
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
	id, err := helpers.ReadIDParam(r)
	if err != nil {
		errs.NotFoundResponse(w, r)
		return
	}

	err = app.models.Workflows.Delete(id)
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
	var input struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	filter := data.Filters{
		Page:         input.Page,
		PageSize:     input.PageSize,
		Sort:         "-id",
		SortSafeList: []string{"id", "name", "-id", "-name"},
	}

	if data.ValidateFilters(v, filter); !v.Valid() {
		errs.FailedValidationResponse(w, r, v.Errors)
		return
	}

	workflows, metadata, err := app.models.Workflows.GetAll(1, filter)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"workflows": workflows, "metadata": metadata}, nil)
}
