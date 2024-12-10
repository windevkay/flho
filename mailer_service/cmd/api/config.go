package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/windevkay/flho/mailer_service/internal/vcs"
)

type application struct {
	config appConfig
	logger *slog.Logger
	server server
	wg     sync.WaitGroup
}

type appConfig struct {
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

var (
	cfg     appConfig
	logger  = slog.New(slog.NewTextHandler(os.Stdout, nil))
	version = vcs.Version()
)

// loadAppConfig loads the application configuration from command-line flags.
// It sets the following configuration options:
// - API server port (default: 4000)
// - Environment (development, staging, production; default: development)
// - SMTP host
// - SMTP port (default: 25)
// - SMTP username
// - SMTP password
// - SMTP sender (default: "FLHO <no-reply@flho.dev>")
// Additionally, it provides a flag to display the version and exit the program.
func loadAppConfig() {
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
}
