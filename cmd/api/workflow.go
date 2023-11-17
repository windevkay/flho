package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/windevkay/flho/internal/data"
)

func (app *application) createWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name            string    `json:"name"`
		States          []string  `json:"states"`
		StartState      string    `json:"startState"`
		EndState        string    `json:"endState"`
		IsTimed         bool      `json:"isTimed"`
		Timeout         time.Time `json:"timeout"`
		CallbackWebhook string    `json:"webhook"`
		Alert           bool      `json:"alert"`
		AlertEmail      string    `json:"alertEmail"`
		AlertWebhook    string    `json:"alertWebhook"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
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
