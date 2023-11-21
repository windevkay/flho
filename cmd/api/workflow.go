package main

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/windevkay/flho/internal/data"
	"github.com/windevkay/flho/internal/validator"
)

func (app *application) createWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name            string    `json:"name"`         // M
		States          []string  `json:"states"`       // M - there should be atleast 2 states (start, end?)
		StartState      string    `json:"startState"`   // M - should match an item in the states slice
		EndState        string    `json:"endState"`     // M - should match an item in the states slice
		CallbackWebhook string    `json:"webhook"`      // O
		IsTimed         bool      `json:"isTimed"`      // M
		Timeout         time.Time `json:"timeout"`      // O - must be provided if isTimed is set to true
		AlertWebhook    string    `json:"alertWebhook"` // O - must be provided if isTimed is set to true
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	v.Check(input.Name != "", "name", "must be provided")
	v.Check(len(input.States) >= 2, "states", "must have atleast 2 values")
	v.Check(validator.Unique(input.States), "states", "must not contain duplicate values")
	v.Check(input.StartState != "", "startState", "must be provided")
	v.Check(slices.Contains(input.States, input.StartState), "startState", "must be part of the states list")
	v.Check(input.EndState != "", "endState", "must be provided")
	v.Check(slices.Contains(input.States, input.EndState), "endState", "must be part of the states list")
	if input.IsTimed {
		v.Check(input.AlertWebhook != "", "alertWebhook", "must be provided")
		// next : validation for timeout
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
