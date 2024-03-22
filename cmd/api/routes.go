package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", app.healthcheckHandler)
	mux.HandleFunc("POST /v1/workflows", app.createWorkflowHandler)
	mux.HandleFunc("GET /v1/workflows/{id}", app.showWorkflowHandler)
	mux.HandleFunc("PATCH /v1/workflows/{id}", app.updateWorkflowHandler)
	mux.HandleFunc("DELETE /v1/workflows/{id}", app.deleteWorkflowHandler)

	return app.recoverPanic(mux)
}
