package main

import (
	"context"
)

func main() {
	loadAppConfig()
	ctx, cancel := context.WithCancel(context.Background())

	app := &application{
		config:    cfg,
		ctx:       ctx,
		cancelCtx: cancel,
		logger:    logger,
	}

	err := app.serveHTTP()
	if err != nil {
		return
	}
}
