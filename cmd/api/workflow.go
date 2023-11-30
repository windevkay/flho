package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/windevkay/flho/internal/data"
	"github.com/windevkay/flho/internal/validator"
)

func (app *application) createWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name            string   `json:"name"`       // M
		States          []string `json:"states"`     // M - there should be atleast 2 states (start, end?)
		StartState      string   `json:"startState"` // M - should match an item in the states slice
		EndState        string   `json:"endState"`   // M - should match an item in the states slice
		CallbackWebhook string   `json:"webhook"`    // O
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	workflow := &data.Workflow{
		Name:            input.Name,
		States:          input.States,
		StartState:      input.StartState,
		EndState:        input.EndState,
		CallbackWebhook: input.CallbackWebhook,
	}

	if data.ValidateWorkflow(v, workflow); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) showWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	workflow := data.Workflow{
		ID:        id,
		CreatedAt: time.Now(),
		Name:      "PRIMARYJOB001",
		Version:   1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"workflow": workflow}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
