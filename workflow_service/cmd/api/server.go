package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/windevkay/flho/workflow_service/internal/queue"
	"github.com/windevkay/flhoutils/helpers"
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
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		app.logger.Info("intercepted signal", "signal", s.String())

		// close message queue channel
		app.mqChannel.Close()

		// begin timed server shutdown
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

// serveQueue listens to the message queue and processes incoming messages.
// It runs in a background goroutine and consumes messages from the specified service queue.
// Depending on the source exchange of the message, it routes the message to the appropriate handler.
// If the source exchange is not recognized, it logs a warning.
//
// The function uses the following parameters for consuming messages:
// - serviceQueue: The name of the queue to consume messages from.
// - autoAck: Automatic acknowledgment mode (set to true).
// - exclusive: Exclusive consumer mode (set to false).
// - noLocal: No local flag (set to false).
// - noWait: No wait flag (set to false).
// - args: Additional arguments for the consume method (set to nil).
//
// If an error occurs while starting to consume messages, it logs the error and returns.
func (app *application) serveQueue() {
	helpers.RunInBackground(func() {
		msgs, err := app.mqChannel.Consume(
			queue.ServiceQueue,
			"",
			true,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			app.logger.Error(fmt.Sprintf("Failed to start consuming messages: %v", err))
			return
		}

		for d := range msgs {
			source := d.Headers["source_exchange"]
			routingKey := d.RoutingKey

			switch source {
			case "workflow_service_exchange":
				// we should really just pass the routing key to an appropriate handler here
				log.Print(routingKey)
			default:
				app.logger.Warn(fmt.Sprintf("Unhandled source exchange: %v", source))
			}
		}
	}, &app.wg)
}
