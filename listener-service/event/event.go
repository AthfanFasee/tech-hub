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

// Create a queue with random name
func declareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"",    // Name
		false, // Durable?
		false, // Delete when unused?
		true,  // Exclusive?
		false, // No-wait?
		nil,   // Any specific arguments?
	)
}
