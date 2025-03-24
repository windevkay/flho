package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type connectFunc func() (any, error)

func connectWithRetry(connect connectFunc, maxRetries int, initialBackoff time.Duration) (any, error) {
	var conn any
	var err error

	for i := range maxRetries {
		conn, err = connect()
		if err == nil {
			return conn, nil
		}

		log.Printf("Failed to connect (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(initialBackoff * (1 << i)) // Exponential backoff
	}

	return nil, err
}

func connectToDB(cfg appConfig) (*mongo.Client, error) {
	connect := func() (any, error) {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.db.connectTimeout)
		defer cancel()

		clientOpts := options.Client().
			ApplyURI(cfg.db.uri).
			SetMaxPoolSize(cfg.db.maxPoolSize)

		return mongo.Connect(ctx, clientOpts)
	}

	conn, err := connectWithRetry(connect, 5, time.Second)
	if err != nil {
		return nil, err
	}

	client := conn.(*mongo.Client)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}
