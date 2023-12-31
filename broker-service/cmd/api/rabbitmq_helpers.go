package main

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/AthfanFasee/broker/event"
)

// Pushes an event to RabbitMQ
func (app *application) pushToQueue(eventName string, eventData any, routingKey string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := RabbitPayload{
		Name: eventName,
		Data: eventData,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), routingKey)
	if err != nil {
		return err
	}

	return nil
}

// Logs error via RabbitMQ
func (app *application) logViaRabbit(name, errorMessage, severity string) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	stackTrace := string(debug.Stack())

	logMessage := fmt.Sprintf("Timestamp: %s\nError: %s\nStackTrace:\n%s", timestamp, errorMessage, stackTrace)

	err := app.pushToQueue(name, logMessage, severity)
	if err != nil {
		return err
	}

	return nil
}
