package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type connectFunc func() (any, error)

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func connectWithRetry(connect connectFunc, maxRetries int, initialBackoff time.Duration) (any, error) {
	var conn any
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = connect()
		if err == nil {
			return conn, nil
		}

		log.Printf("Failed to connect (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(initialBackoff * (1 << i)) // Exponential backoff
	}

	return nil, err
}

func connectToMessageQueue(cfg config) (*amqp.Connection, error) {
	connect := func() (any, error) {
		return amqp.Dial(cfg.messageQueueAddr)
	}

	conn, err := connectWithRetry(connect, 5, time.Second)
	if err != nil {
		return nil, err
	}

	return conn.(*amqp.Connection), nil
}

func connectToMailerServer(cfg config) (*grpc.ClientConn, error) {
	connect := func() (any, error) {
		return grpc.NewClient(cfg.mailerServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := connectWithRetry(connect, 5, time.Second)
	if err != nil {
		return nil, err
	}

	return conn.(*grpc.ClientConn), nil
}
