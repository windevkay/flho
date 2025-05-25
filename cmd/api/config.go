package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Netflix/go-env"
	"github.com/rs/zerolog"
)

type application struct {
	config    appConfig
	ctx       context.Context
	cancelCtx context.CancelFunc
	logger    zerolog.Logger
	wg        sync.WaitGroup
}

type appConfig struct {
	Env      string `env:"ENVIRONMENT,default=development"`
	Extras   env.EnvSet
	HttpPort int `env:"HTTP_PORT,default=4000"`
	Log      struct {
		Level  string `env:"LOG_LEVEL,default=info"`
		Format string `env:"LOG_FORMAT,default=json"`
	}
}

var (
	cfg    appConfig
	logger zerolog.Logger
)

func setupLogger() {
	// Set global time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Set the global log level
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		// Default to info level if parsing fails
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output format based on environment
	var output zerolog.ConsoleWriter

	// Pretty logging for development
	if cfg.Env == "development" && cfg.Log.Format == "pretty" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		// Initialize the logger with a console writer
		logger = zerolog.New(output).
			With().
			Timestamp().
			Str("app", "flho").
			Str("env", cfg.Env).
			Logger()
	} else {
		// Initialize the logger with standard JSON output
		logger = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Str("app", "flho").
			Str("env", cfg.Env).
			Logger()
	}
}

func loadAppConfig() {
	es, err := env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	cfg.Extras = es

	// Initialize the logger after config is loaded
	setupLogger()
}
