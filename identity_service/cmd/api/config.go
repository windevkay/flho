package main

import (
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/windevkay/flho/identity_service/internal/data"
	"github.com/windevkay/flho/identity_service/internal/queue"
	"github.com/windevkay/flho/identity_service/internal/rpc"
	"github.com/windevkay/flho/identity_service/internal/services"
	"github.com/windevkay/flho/identity_service/internal/vcs"
)

type application struct {
	config    appConfig
	logger    *slog.Logger
	mqChannel *amqp.Channel
	models    data.Models
	rpc       rpc.Clients
	wg        sync.WaitGroup
}

type appConfig struct {
	port              int
	env               string
	mailerServiceAddr string
	messageQueueAddr  string
	db                struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type appConnections struct {
	db          *sql.DB
	amqp        *amqp.Connection
	rpc         rpc.Connections
	amqpChannel *amqp.Channel
}

var (
	cfg                   appConfig
	logger                = slog.New(slog.NewTextHandler(os.Stdout, nil))
	identityServiceConfig *services.IdentityServiceConfig
	version               = vcs.Version()
)

// loadAppConfig loads the application configuration from command-line flags.
// It sets various configuration options such as server port, environment, gRPC server addresses,
// message queue address, database connection settings, and rate limiter settings.
// It also provides an option to display the application version and exit.
//
// Flags:
// -port: API server port (default: 4000)
// -env: Environment (development|staging|production) (default: "development")
// -mailer-server: Mailer service address
// -message-queue-server: Message queue address
// -db-dsn: PostgreSQL DSN
// -db-max-open-conns: PostgreSQL max open connections (default: 25)
// -db-max-idle-conns: PostgreSQL max idle connections (default: 25)
// -db-max-idle-time: PostgreSQL max connection idle time (default: 15m)
// -limiter-rps: Rate limiter maximum requests per second (default: 2)
// -limiter-burst: Rate limiter maximum burst (default: 4)
// -limiter-enabled: Enable rate limiter (default: true)
// -version: Display version and exit
func loadAppConfig() {
	// environment flags
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// gRPC servers
	flag.StringVar(&cfg.mailerServiceAddr, "mailer-server", "", "Mailer service address")

	// message queue connection string
	flag.StringVar(&cfg.messageQueueAddr, "message-queue-server", "", "Message queue address")

	// db connection and db pool settings flags
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	// rate limiter flags
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}
}

// setupConnections initializes and establishes connections to various services required by the application.
// It connects to the primary datastore (PostgreSQL), gRPC servers (e.g., mailer server), and the message queue (RabbitMQ).
// It also sets up message exchanges and queues for RabbitMQ.
//
// Returns:
//   - *appConnections: A struct containing the established connections.
//   - error: An error if any of the connections fail to be established.
func setupConnections() (*appConnections, error) {
	// connect to primary datastore - postgres
	db, err := openDB(cfg)
	if err != nil {
		return nil, err
	}

	logger.Info("Connection to DB has been established")
	// gRPC server connections
	// mailer server
	mailerConn, err := connectToMailerServer(cfg)
	if err != nil {
		return nil, err
	}

	logger.Info("gRPC connections have been established")
	// connect to message queue - rabbitmq
	amqpConn, err := connectToMessageQueue(cfg)
	if err != nil {
		return nil, err
	}

	ch, err := amqpConn.Channel()
	if err != nil {
		return nil, err
	}
	// setup message exchanges and queues
	err = queue.SetupExchanges(ch)
	if err != nil {
		return nil, err
	}

	logger.Info("AMQP connection has been established")

	connections := &appConnections{
		db:   db,
		amqp: amqpConn,
		rpc: rpc.Connections{
			MailerConn: mailerConn,
		},
		amqpChannel: ch,
	}

	return connections, nil
}

// publishMetrics publishes various application metrics using the expvar package.
// It publishes the following metrics:
// - "version": the current version of the application.
// - "goroutines": the number of active goroutines.
// - "database": the database connection statistics.
// - "timestamp": the current Unix timestamp.
//
// Parameters:
// - db: a pointer to the sql.DB instance representing the database connection.
func publishMetrics(db *sql.DB) {
	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))
}

// registerServiceConfigs initializes and registers the configuration settings
// for various services. It sets up the necessary dependencies such as models,
// RPC clients, message queue channel, wait group, and logger for the service.
func (app *application) registerServiceConfigs() {
	// register identity service configs
	identityServiceConfig = &services.IdentityServiceConfig{
		Models:    app.models,
		Rpclients: app.rpc,
		Channel:   app.mqChannel,
		Wg:        &app.wg,
		Logger:    app.logger,
	}
}
