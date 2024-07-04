package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/windevkay/flho/internal/data"
	errs "github.com/windevkay/flhoutils/errors"
	"github.com/windevkay/flhoutils/helpers"
	"github.com/windevkay/flhoutils/validator"
)

func (app *application) createWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name         string       `json:"name"`
		States       []string     `json:"states"`
		StartState   string       `json:"startState"`
		EndState     string       `json:"endState"`
		RetryWebhook string       `json:"retryWebhook"`
		RetryAfter   data.Timeout `json:"retryAfter"`
	}

	err := helpers.ReadJSON(w, r, &input)
	if err != nil {
		errs.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	workflow := &data.Workflow{
		UniqueID:     helpers.GenerateUniqueId(15),
		Name:         input.Name,
		States:       input.States,
		StartState:   input.StartState,
		EndState:     input.EndState,
		RetryWebhook: input.RetryWebhook,
		RetryAfter:   input.RetryAfter,
		Active:       true,
	}

	if data.ValidateWorkflow(v, workflow); !v.Valid() {
		errs.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Workflows.Insert(workflow)
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
	Name         *string       `json:"name"`
	States       []string      `json:"states"`
	StartState   *string       `json:"startState"`
	EndState     *string       `json:"endState"`
	RetryWebhook *string       `json:"retryWebhook"`
	RetryAfter   *data.Timeout `json:"retryAfter"`
}) {
	if input.Name != nil {
		workflow.Name = *input.Name
	}
	if input.States != nil {
		workflow.States = input.States
	}
	if input.StartState != nil {
		workflow.StartState = *input.StartState
	}
	if input.EndState != nil {
		workflow.EndState = *input.EndState
	}
	if input.RetryWebhook != nil {
		workflow.RetryWebhook = *input.RetryWebhook
	}
	if input.RetryAfter != nil {
		workflow.RetryAfter = *input.RetryAfter
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
		Name         *string       `json:"name"`
		States       []string      `json:"states"`
		StartState   *string       `json:"startState"`
		EndState     *string       `json:"endState"`
		RetryWebhook *string       `json:"retryWebhook"`
		RetryAfter   *data.Timeout `json:"retryAfter"`
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
		Name   string
		States []string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = helpers.ReadString(qs, "name", "")
	input.States = helpers.ReadCSV(qs, "states", []string{})
	input.Filters.Page = helpers.ReadInt(qs, "page", 1, v)
	input.Filters.PageSize = helpers.ReadInt(qs, "page_size", 20, v)
	input.Filters.Sort = helpers.ReadString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "name", "-id", "-name"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		errs.FailedValidationResponse(w, r, v.Errors)
		return
	}

	workflows, metadata, err := app.models.Workflows.GetAll(input.Name, input.States, input.Filters)
	if err != nil {
		errs.ServerErrorResponse(w, r, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, helpers.Envelope{"workflows": workflows, "metadata": metadata}, nil)
}
