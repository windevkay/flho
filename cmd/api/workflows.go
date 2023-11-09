package main

import (
	"fmt"
	"net/http"
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

	fmt.Fprintf(w, "show the details of workflow %d\n", id)
}