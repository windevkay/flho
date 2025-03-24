package main

import (
	"context"
	"os"

	"github.com/windevkay/flho/internal/data"
	"github.com/windevkay/flho/internal/mailer"
	"github.com/windevkay/flhoutils/helpers"
)

func main() {
	app, connections := createApp()
	defer cleanup(connections)

	publishMetrics(connections.db)

	// start http server
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

	app := &application{
		config:         cfg,
		logger:         logger,
		mailer:         mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		models:         data.GetModels(connections.db, cfg.db.database),
		backgroundFunc: helpers.RunInBackground,
	}

	// register service configs
	app.registerServiceConfigs()

	return app, connections
}

func cleanup(connections *appConnections) {
	connections.db.Disconnect(context.Background())
}
