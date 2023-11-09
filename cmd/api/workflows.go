package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/windevkay/flho/internal/data"
)

func (app *application) createWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new workflow")
}

func (app *application) showWorkflowHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	workflow := data.Workflow{
		ID: id,
		CreatedAt: time.Now(),
		Name: "PRIMARYJOB001",
		Version: 1,
	}

	err = app.writeJSON(w, http.StatusOK, workflow, nil)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The server encounterd a problem and could not process your request", http.StatusInternalServerError)
	}
}
