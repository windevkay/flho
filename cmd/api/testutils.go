package main

import (
	"io"
	"log/slog"

	"github.com/windevkay/flho/internal/data"
)

func newTestApplication() *application {
	return &application{
		config: config{env: "test"},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		models: data.GetMockModels(),
	}
}
