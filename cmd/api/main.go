package main

import (
	"context"
	"os"

	"github.com/windevkay/flho/internal/data"
	"github.com/windevkay/flhoutils/helpers"
)

func main() {
	app, connections := createApp()
	defer connections.db.Disconnect(app.ctx)

	publishMetrics(connections.db)

	err := app.serveHTTP()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func createApp() (*application, *appConnections) {
	loadAppConfig()

	connections, err := setupConnections()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	app := &application{
		config:    cfg,
		ctx:       ctx,
		cancelCtx: cancel,
		logger:    logger,
		models:    data.GetModels(connections.db, cfg.db.database),
		bg:        helpers.RunInBackground,
	}

	app.registerServiceConfigs()

	return app, connections
}
