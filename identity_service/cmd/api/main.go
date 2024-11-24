package main

import (
	"os"

	_ "github.com/lib/pq"
	"github.com/windevkay/flho/identity_service/internal/data"
	"github.com/windevkay/flho/identity_service/internal/rpc"
)

func main() {
	loadAppConfig()

	connections, err := setupConnections()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer connections.db.Close()
	defer connections.rpc.Close()
	defer connections.amqp.Close()

	publishMetrics(connections.db)

	app := &application{
		config:    cfg,
		logger:    logger,
		models:    data.GetModels(connections.db),
		rpc:       rpc.GetClients(connections.rpc),
		mqChannel: connections.amqpChannel,
	}

	// register service configs
	app.registerServiceConfigs()

	// listen for messages from event queue
	app.serveQueue()

	// start http server
	err = app.serveHTTP()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
