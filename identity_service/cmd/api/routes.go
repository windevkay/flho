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
	// create and activate user
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerIdentityHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activate", app.activateIdentityHandler)
	// get a token
	router.HandlerFunc(http.MethodPost, "/v1/auth/token", app.createAuthenticationTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.rateLimit(router)))
}
