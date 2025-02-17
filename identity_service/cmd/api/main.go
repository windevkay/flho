package main

import (
	"context"
	"os"

	"github.com/windevkay/flho/identity_service/docs"
	"github.com/windevkay/flho/identity_service/internal/data"
	"github.com/windevkay/flho/identity_service/internal/rpc"
)

func main() {
	// setup swagger/OpenAPI docs
	setupDocs()

	app, connections := createApp()

	defer connections.db.Disconnect(context.Background())
	defer connections.rpc.Close()
	defer connections.amqp.Close()

	publishMetrics(connections.db)

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

	// register service configs
	app.registerServiceConfigs()

	return app, connections
}

func setupDocs() {
	docs.SwaggerInfo.Title = "Identity Service API"
	docs.SwaggerInfo.Description = "This is the API for the Identity Service."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:4002"
	docs.SwaggerInfo.BasePath = "/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
}
