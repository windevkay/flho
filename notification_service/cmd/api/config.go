package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"

	"github.com/windevkay/flho/notification_service/internal/vcs"
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

// loadAppConfig loads the application configuration from environment variables.
// It sets the following configuration options:
// - API server port (default: 4000)
// - Environment (development, staging, production; default: development)
// - SMTP host
// - SMTP port (default: 25)
// - SMTP username
// - SMTP password
// - SMTP sender (default: "FLHO <no-reply@flho.dev>")
// Additionally, it provides an option to display the version and exit the program.
func loadAppConfig() {
	cfg.port = getEnvAsInt("PORT", 4000)
	cfg.env = getEnv("ENV", "development")
	cfg.smtp.host = getEnv("SMTP_HOST", "")
	cfg.smtp.port = getEnvAsInt("SMTP_PORT", 25)
	cfg.smtp.username = getEnv("SMTP_USERNAME", "")
	cfg.smtp.password = getEnv("SMTP_PASSWORD", "")
	cfg.smtp.sender = getEnv("SMTP_SENDER", "FLHO <no-reply@flho.dev>")

	if getEnvAsBool("VERSION", false) {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}
}

// getEnv reads an environment variable or returns a default value if not set.
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as an integer or returns a default value if not set.
func getEnvAsInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsBool reads an environment variable as a boolean or returns a default value if not set.
func getEnvAsBool(name string, defaultValue bool) bool {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
