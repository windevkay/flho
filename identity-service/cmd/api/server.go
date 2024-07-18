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

func (app *application) serve() error {
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
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		app.logger.Info("intercepted signal", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("...finishing background tasks", "addr", srv.Addr)

		app.wg.Wait()
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
