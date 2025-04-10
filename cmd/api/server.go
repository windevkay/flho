package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// serveHTTP starts the HTTP server and handles graceful shutdown on receiving
// termination signals (SIGINT, SIGTERM). It configures the server with timeouts
// and an error logger, and waits for background tasks to complete before fully
// shutting down. The function returns an error if the server fails to start or
// if there are issues during the shutdown process.
func (app *application) serveHTTP() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)
		s := <-quit
		app.logger.Info("intercepted signal", "signal", s.String())

		// begin timed server shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("...finishing background tasks", "addr", srv.Addr)

		app.wg.Wait()
		app.cancelCtx()
		shutdownError <- nil
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	err := srv.ListenAndServe()
	// check if error is not a result of calling Shutdown
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	// check for errors in shutting down
	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("server stopped", "addr", srv.Addr)

	return nil
}
