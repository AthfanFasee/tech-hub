package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// Creates an exchange
func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"tech_hub", // Name
		"topic",    // Type
		true,       // Is it durable?
		false,      // Auto-Deleted?
		false,      // Internal?
		false,      // No-wait?
		nil,        // Any specific arguments?
	)
}
