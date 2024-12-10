package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ServiceExchange string = "workflow_service_exchange"
	ServiceQueue    string = "workflow_service_queue"
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
	err := exchange(ch, ServiceExchange)
	if err != nil {
		return err
	}
	return nil
}

// setupExternalExchanges sets up the external exchanges and binds the service queue to the events
// from these exchanges. It declares the external exchanges and creates queue bindings for each event
// specified in the externalExchanges list.
//
// Parameters:
// - ch: The AMQP channel used to declare exchanges and bind queues.
//
// Returns:
// - error: An error if any exchange declaration or queue binding fails, otherwise nil.
func setupExternalExchanges(ch *amqp.Channel) error {
	q, err := ch.QueueDeclare(ServiceQueue, false, false, true, false, nil)
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

func SetupExchanges(ch *amqp.Channel) error {
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
//	action: The action related to the message. e.g create | update | delete
//
// Returns:
//
//	error: An error if the message could not be serialized or published, otherwise nil.
func SendMessage(ch *amqp.Channel, message any, entityName string, action string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	err = ch.PublishWithContext(
		ctx,
		ServiceExchange,
		fmt.Sprintf("%s.%s.%s", ServiceExchange, entityName, action), // routing key e.g. identity_service_exchange.token.create
		false,
		false,
		amqp.Publishing{
			Headers: amqp.Table{
				"source_exchange": ServiceExchange,
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
