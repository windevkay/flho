package main

import (
	"context"
	"os"

	"github.com/windevkay/flho/identity_service/internal/data"
	"github.com/windevkay/flho/identity_service/internal/rpc"
)

func main() {
	app, connections := createApp()

	defer connections.db.Disconnect(context.Background())
	defer connections.rpc.Close()
	defer connections.amqp.Close()

	publishMetrics(connections.db)

	// register service configs
	app.registerServiceConfigs()

	// listen for messages from event queue
	app.serveQueue()

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
		config:    cfg,
		logger:    logger,
		models:    data.GetModels(connections.db, cfg.db.database),
		rpc:       rpc.GetClients(connections.rpc),
		mqChannel: connections.amqpChannel,
	}

	return app, connections
}
