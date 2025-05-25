package main

import (
	"github.com/windevkay/flho/internal/handlers"
	"github.com/windevkay/flho/pkg/apperrors"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(apperrors.NotFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(apperrors.MethodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/healthcheck", handlers.HealthcheckHandler)

	return app.recoverPanic(app.contextWithLogger(router))
}
