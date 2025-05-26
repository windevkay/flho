package main

import (
	"context"
	"github.com/Netflix/go-env"
	"github.com/rs/zerolog"
	"log"
	"os"
	"sync"
	"time"
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

func main() {
	var cfg appConfig
	es, err := env.UnmarshalFromEnviron(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Extras = es

	var logger zerolog.Logger
	zerolog.TimeFieldFormat = time.RFC3339
	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("app", "flho").
		Str("env", cfg.Env).
		Logger()

	ctx, cancel := context.WithCancel(context.Background())

	app := &application{
		config:    cfg,
		ctx:       ctx,
		cancelCtx: cancel,
		logger:    logger,
	}

	err = app.serveHTTP()
	if err != nil {
		return
	}
}
