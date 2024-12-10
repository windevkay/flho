package main

import (
	"os"

	"github.com/windevkay/flho/mailer_service/internal/mailer"
	pb "github.com/windevkay/flho/mailer_service/proto"

	_ "github.com/lib/pq"
)

type server struct {
	pb.UnimplementedMailerServer
	mailer mailer.Mailer
}

func main() {
	loadAppConfig()

	app := &application{
		config: cfg,
		logger: logger,
		server: server{
			mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		},
	}

	err := app.serveGRPC()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
