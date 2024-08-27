package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/windevkay/flho/mailer_service/internal/mailer"
	"github.com/windevkay/flho/mailer_service/internal/vcs"
	pb "github.com/windevkay/flho/mailer_service/proto"

	_ "github.com/lib/pq"
)

var (
	version = vcs.Version()
)

type config struct {
	port int
	env  string
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type server struct {
	pb.UnimplementedMailerServer
	mailer mailer.Mailer
}

type application struct {
	config config
	logger *slog.Logger
	server server
	wg     sync.WaitGroup
}

func main() {
	var cfg config
	// environment flags
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// smtp flags
	flag.StringVar(&cfg.smtp.host, "smtp-host", "", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "FLHO <no-reply@flho.dev>", "SMTP sender")

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	// provide a structured logger for the application
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := &application{
		config: cfg,
		logger: logger,
		server: server{
			mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		},
	}

	err := app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
