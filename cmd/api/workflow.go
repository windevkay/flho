package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/windevkay/flho/internal/data"
	"github.com/windevkay/flho/internal/validator"
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

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	workflow := &data.Workflow{
		UniqueID:     app.generateWorkflowUniqueId(),
		Name:         input.Name,
		States:       input.States,
		StartState:   input.StartState,
		EndState:     input.EndState,
		RetryWebhook: input.RetryWebhook,
		RetryAfter:   input.RetryAfter,
		Active:       true,
	}

	if data.ValidateWorkflow(v, workflow); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Workflows.Insert(workflow)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// handy resource location header for clients
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/workflows/%d", workflow.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"workflow": workflow}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	workflow, err := app.models.Workflows.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"workflow": workflow}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	workflow, err := app.models.Workflows.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
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

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// achieve partial updates using non nil values
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

	v := validator.New()

	if data.ValidateWorkflow(v, workflow); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Workflows.Update(workflow)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"workflow": workflow}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Workflows.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "success: the workflow has been deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name   string
		States []string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	input.States = app.readCSV(qs, "states", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "name", "-id", "-name"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}
