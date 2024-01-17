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
		Name                        string       `json:"name"`
		States                      []string     `json:"states"`
		StartState                  string       `json:"startState"`
		EndState                    string       `json:"endState"`
		CallbackWebhook             string       `json:"webhook,omitempty"`
		Retry                       bool         `json:"retry"`
		RetryAfter                  data.Timeout `json:"retryAfter,omitempty"`
		RetryURL                    string       `json:"retryUrl,omitempty"`
		CircuitBreaker              bool         `json:"circuitBreaker"`
		CircuitBreakerFailureCount  int32        `json:"circuitBreakerFailureCount,omitempty"`
		CircuitBreakerOpenTimeout   data.Timeout `json:"circuitBreakerOpenTimeout,omitempty"`
		CircuitBreakerHalfOpenCount int32        `json:"circuitBreakerHalfOpenCount,omitempty"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	workflow := &data.Workflow{
		UniqueID:                    app.generateWorkflowUniqueId(),
		Name:                        input.Name,
		States:                      input.States,
		StartState:                  input.StartState,
		EndState:                    input.EndState,
		CallbackWebhook:             input.CallbackWebhook,
		Retry:                       input.Retry,
		RetryAfter:                  input.RetryAfter,
		RetryURL:                    input.RetryURL,
		CircuitBreaker:              input.CircuitBreaker,
		CircuitBreakerStatus:        "CLOSED",
		CircuitBreakerFailureCount:  input.CircuitBreakerFailureCount,
		CircuitBreakerOpenTimeout:   input.CircuitBreakerOpenTimeout,
		CircuitBreakerHalfOpenCount: input.CircuitBreakerHalfOpenCount,
		Active:                      true,
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
