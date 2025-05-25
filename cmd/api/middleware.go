package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/windevkay/flho/pkg/apperrors"
	"github.com/windevkay/flho/pkg/utils"
	"net/http"
)

func (app *application) contextWithLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqId := uuid.New().String()
		reqLogger := app.logger.With().
			Str("request_id", reqId).
			Str("remote_addr", r.RemoteAddr).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Logger()

		ctx := utils.WithLogger(r.Context(), reqLogger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				apperrors.ServerErrorResponse(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
