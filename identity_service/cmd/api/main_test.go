package main

import (
	"log"
	"os"
	"sync"
	"testing"

	"github.com/joho/godotenv"
	"github.com/rabbitmq/amqp091-go"
)

var testApp *application

func TestMain(m *testing.M) {
	// load test environment variables
	err := godotenv.Load("../../.env.testing")
	if err != nil {
		log.Fatal("Error loading .env.testing file")
	}

	app, connections := createApp()
	defer cleanup(connections)

	testApp = app
	// mock message queue and background ops functions on service config
	serviceConfig.Message = func(ch *amqp091.Channel, data interface{}, entity string, action string) error {
		return nil
	}
	serviceConfig.Background = func(f func(), wg *sync.WaitGroup) {}

	code := m.Run()

	os.Exit(code)
}
