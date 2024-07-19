package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
	errs "github.com/windevkay/flhoutils/errors"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(errs.NotFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(errs.MethodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/workflows", app.listWorkflowHandler)
	router.HandlerFunc(http.MethodPost, "/v1/workflows", app.createWorkflowHandler)
	router.HandlerFunc(http.MethodGet, "/v1/workflows/:id", app.showWorkflowHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/workflows/:id", app.updateWorkflowHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/workflows/:id", app.deleteWorkflowHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.rateLimit(app.authenticate(router))))
}
