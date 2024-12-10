package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/windevkay/flho/mailer_service/proto"
	"google.golang.org/grpc"
)

func (app *application) serveGRPC() error {

	srv, err := net.Listen("tcp", fmt.Sprintf(":%d", app.config.port))
	if err != nil {
		return err
	}

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit
		app.logger.Info("intercepted signal", "signal", s.String())

		err = srv.Close()
		if err != nil {
			shutdownError <- err
		}

		app.logger.Info("...finishing background tasks", "addr", srv.Addr)

		app.wg.Wait()
		shutdownError <- nil
	}()

	s := grpc.NewServer()
	pb.RegisterMailerServer(s, &app.server)

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	err = s.Serve(srv)
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
