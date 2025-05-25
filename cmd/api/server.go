package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

// serveHTTP starts and manages the lifecycle of the HTTP server, including graceful shutdown on termination signals.
func (app *application) serveHTTP() error {
	const ReadTimeout int = 5
	const WriteTimeout int = 10
	const ShutdownTimeout int = 30

	// Create a logger adapter that satisfies the standard log.Logger interface
	zerologAdapter := zerolog.New(os.Stderr).Level(zerolog.ErrorLevel)
	stdLogger := log.New(zerologAdapter, "", 0)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.HttpPort),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Duration(ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(WriteTimeout) * time.Second,
		ErrorLog:     stdLogger,
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGQUIT)
		s := <-quit
		app.logger.Info().Str("signal", s.String()).Msg("intercepted signal")

		// begin timed server shutdown
		ctx, cancel := context.WithTimeout(app.ctx, time.Duration(ShutdownTimeout)*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info().Str("addr", srv.Addr).Msg("...finishing background tasks")

		app.wg.Wait()
		app.cancelCtx()

		shutdownError <- nil
	}()

	app.logger.Info().Str("addr", srv.Addr).Str("env", app.config.Env).Msg("starting server")

	err := srv.ListenAndServe()
	// check if the error is not a result of calling Shutdown
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	// check for errors in shutting down
	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info().Str("addr", srv.Addr).Msg("server stopped")

	return nil
}
