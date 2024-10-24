package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/windevkay/flhoutils/helpers"
)

const (
	serviceExchange = "identity_service_exchange"
	serviceQueue    = "identity_service_queue"
)

type externalServiceData struct {
	name   string
	events []struct{ entityName string }
}

var (
	// list of other exchanges this service is interested in setting up
	externalExchanges = []externalServiceData{
		{
			name: "workflow_service_exchange",
			events: []struct{ entityName string }{
				{
					entityName: "workflow",
				},
			},
		},
	}
)

func exchange(ch *amqp.Channel, exchangeName string) error {
	err := ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	return nil
}

func setupServiceExchange(ch *amqp.Channel) error {
	err := exchange(ch, serviceExchange)
	if err != nil {
		return err
	}
	return nil
}

func setupExternalExchanges(ch *amqp.Channel) error {
	q, err := ch.QueueDeclare(serviceQueue, false, false, true, false, nil)
	if err != nil {
		return err
	}

	for i := 0; i < len(externalExchanges); i++ {
		err := exchange(ch, externalExchanges[i].name)

		if err != nil {
			return err
		}
		// create queue bindings
		events := externalExchanges[i].events
		for j := 0; j < len(events); j++ {
			err = ch.QueueBind(
				q.Name,
				fmt.Sprintf("%s.%s.*", externalExchanges[i].name, events[j].entityName), // routing key
				externalExchanges[i].name,
				false,
				nil,
			)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func setupMessageQueues(ch *amqp.Channel) error {
	err := setupServiceExchange(ch)
	if err != nil {
		return err
	}
	err = setupExternalExchanges(ch)
	if err != nil {
		return err
	}
	return nil
}

// sendQueueMessage sends a message to the specified AMQP channel with a given entity and action.
// The message is serialized to JSON format before being published.
//
// Type Parameters:
//
//	T: The type of the message to be sent.
//
// Parameters:
//
//	ch: The AMQP channel to publish the message to.
//	message: The message to be sent.
//	entityName: The name of the entity related to the message.
//	action: The action related to the message.
//
// Returns:
//
//	error: An error if the message could not be serialized or published, otherwise nil.
func (app *application) sendQueueMessage(message any, entityName string, action string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	err = app.mqChannel.PublishWithContext(
		ctx,
		serviceExchange,
		fmt.Sprintf("%s.%s.%s", serviceExchange, entityName, action), // routing key e.g. identity_service_exchange.token.create
		false,
		false,
		amqp.Publishing{
			Headers: amqp.Table{
				"source_exchange": serviceExchange,
			},
			ContentType: "text/plain",
			Body:        body,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (app *application) listenToMsgQueue() error {
	msgs, err := app.mqChannel.Consume(
		serviceQueue,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	helpers.RunInBackground(func() {
		for d := range msgs {
			log.Println(d.Body)
		}
	}, &app.wg)

	return nil
}
