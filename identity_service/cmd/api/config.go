package main

import (
	"context"
	"expvar"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/windevkay/flho/identity_service/internal/data"
	"github.com/windevkay/flho/identity_service/internal/queue"
	"github.com/windevkay/flho/identity_service/internal/rpc"
	"github.com/windevkay/flho/identity_service/internal/services"
	"github.com/windevkay/flho/identity_service/internal/vcs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type application struct {
	config         appConfig
	logger         *slog.Logger
	mqChannel      *amqp.Channel
	messageFunc    services.SendMessageToQueueFunc
	backgroundFunc services.RunInBackgroundFunc
	models         data.Models
	rpc            rpc.Clients
	wg             sync.WaitGroup
}

type appConfig struct {
	port              int
	env               string
	mailerServiceAddr string
	messageQueueAddr  string
	db                struct {
		uri            string
		database       string
		maxPoolSize    uint64
		connectTimeout time.Duration
	}
	jwt struct {
		secret string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

type appConnections struct {
	db          *mongo.Client
	amqp        *amqp.Connection
	rpc         rpc.Connections
	amqpChannel *amqp.Channel
}

var (
	cfg           appConfig
	logger        = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	serviceConfig services.ServiceConfig
	version       = vcs.Version()
)

// loadAppConfig loads the application configuration from environment variables.
// It sets various configuration options such as server port, environment, gRPC server addresses,
// message queue address, database connection settings, and rate limiter settings.
// It also provides an option to display the application version and exit.
//
// Environment Variables:
// - PORT: API server port (default: 4000)
// - ENV: Environment (development|staging|production) (default: "development")
// - MAILER_SERVER: Mailer service address
// - MESSAGE_QUEUE_SERVER: Message queue address
// - DB_URI: DB connection URI
// - DB_NAME: Database name
// - DB_MAX_POOL_SIZE: DB max connection pool size (default: 100)
// - DB_CONNECT_TIMEOUT: DB connection timeout (default: 10s)
// - LIMITER_RPS: Rate limiter maximum requests per second (default: 2)
// - LIMITER_BURST: Rate limiter maximum burst (default: 4)
// - LIMITER_ENABLED: Enable rate limiter (default: true)
func loadAppConfig() {
	cfg.port = getEnvAsInt("PORT", 4000)
	cfg.env = getEnv("ENV", "development")
	cfg.mailerServiceAddr = getEnv("MAILER_SERVER", "")
	cfg.messageQueueAddr = getEnv("MESSAGE_QUEUE_SERVER", "")
	cfg.db.uri = getEnv("DB_URI", "")
	cfg.db.database = getEnv("DB_NAME", "")
	cfg.db.maxPoolSize = getEnvAsUint64("DB_MAX_POOL_SIZE", 100)
	cfg.db.connectTimeout = getEnvAsDuration("DB_CONNECT_TIMEOUT", 10*time.Second)
	cfg.jwt.secret = getEnv("JWT_SECRET", "")
	cfg.limiter.rps = getEnvAsFloat64("LIMITER_RPS", 2)
	cfg.limiter.burst = getEnvAsInt("LIMITER_BURST", 4)
	cfg.limiter.enabled = getEnvAsBool("LIMITER_ENABLED", true)
}

// getEnv reads an environment variable or returns a default value if not set.
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as an integer or returns a default value if not set.
func getEnvAsInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsUint64 reads an environment variable as a uint64 or returns a default value if not set.
func getEnvAsUint64(name string, defaultValue uint64) uint64 {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseUint(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsFloat64 reads an environment variable as a float64 or returns a default value if not set.
func getEnvAsFloat64(name string, defaultValue float64) float64 {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsBool reads an environment variable as a boolean or returns a default value if not set.
func getEnvAsBool(name string, defaultValue bool) bool {
	valueStr := getEnv(name, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// getEnvAsDuration reads an environment variable as a time.Duration or returns a default value if not set.
func getEnvAsDuration(name string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(name, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// setupConnections initializes and establishes connections to various services required by the application.
// It connects to the primary datastore (MongoDB), gRPC servers (e.g., mailer server), and the message queue (RabbitMQ).
// It also sets up message exchanges and queues for RabbitMQ.
//
// Returns:
//   - *appConnections: A struct containing the established connections.
//   - error: An error if any of the connections fail to be established.
func setupConnections() (*appConnections, error) {
	// connect to DB
	mongoClient, err := connectToDB(cfg)
	if err != nil {
		return nil, err
	}

	logger.Info("DB connection has been established")
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
		db:   mongoClient,
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
func publishMetrics(db *mongo.Client) {
	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("db-stats", expvar.Func(func() any {
		result := db.Database(cfg.db.database).RunCommand(context.Background(), bson.D{{Key: "serverStatus", Value: 1}})
		stats, err := result.Raw()
		if err != nil {
			return err.Error()
		}
		return stats.String()
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))
}

// registerServiceConfigs initializes and registers the configuration settings
// for various services. It sets up the necessary dependencies such as models,
// RPC clients, message queue channel, wait group, and logger for the service.
func (app *application) registerServiceConfigs() {
	serviceConfig.Register(app.models, app.rpc, app.mqChannel, &app.wg, app.logger, app.backgroundFunc, app.messageFunc)
}
