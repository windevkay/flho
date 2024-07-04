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

	router.HandlerFunc(http.MethodGet, "/v1/workflows", app.requireAuthenticatedUser(app.listWorkflowHandler))
	router.HandlerFunc(http.MethodPost, "/v1/workflows", app.requireActivatedUser(app.createWorkflowHandler))
	router.HandlerFunc(http.MethodGet, "/v1/workflows/:id", app.requireActivatedUser(app.showWorkflowHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/workflows/:id", app.requireActivatedUser(app.updateWorkflowHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/workflows/:id", app.requireActivatedUser(app.deleteWorkflowHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/auth/token", app.createAuthenticationTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.rateLimit(app.authenticate(router))))
}
