package main

import (
	"expvar"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/windevkay/flho/mailer_service/proto"

	_ "github.com/lib/pq"
	"github.com/windevkay/flho/identity_service/internal/data"
	"github.com/windevkay/flho/identity_service/internal/vcs"
)

var (
	version = vcs.Version()
)

type config struct {
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

type application struct {
	config       config
	logger       *slog.Logger
	mqChannel    *amqp.Channel
	models       data.Models
	mailerClient pb.MailerClient
	wg           sync.WaitGroup
}

func main() {
	var cfg config
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

	// provide a structured logger for the application
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// connect to primary datastore - postgres
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("Connection to DB has been established")

	// gRPC server connections
	// mailer server
	mailerConn, err := connectToMailerServer(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer mailerConn.Close()

	mailerClient := pb.NewMailerClient(mailerConn)

	logger.Info("gRPC connections have been established")

	// connect to message queue - rabbitmq
	amqpConn, err := connectToMessageQueue(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer amqpConn.Close()

	ch, err := amqpConn.Channel()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// setup message exchanges and queues
	err = setupMessageQueues(ch)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("AMQP connection has been established")

	// metrics
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

	app := &application{
		config:       cfg,
		logger:       logger,
		models:       data.GetModels(db),
		mailerClient: mailerClient,
		mqChannel:    ch,
	}

	// listen for messages from event queue
	err = app.listenToMsgQueue()
	if err != nil {
		logger.Error(err.Error())
	}

	// start http server
	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
