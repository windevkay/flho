package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/workflows", app.listWorkflowHandler)
	router.HandlerFunc(http.MethodPost, "/v1/workflows", app.createWorkflowHandler)
	router.HandlerFunc(http.MethodGet, "/v1/workflows/:id", app.showWorkflowHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/workflows/:id", app.updateWorkflowHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/workflows/:id", app.deleteWorkflowHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	return app.recoverPanic(app.rateLimit(router))
}
