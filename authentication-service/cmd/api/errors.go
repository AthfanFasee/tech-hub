package main

import (
	"log"
	"strings"

	"github.com/AthfanFasee/authentication/internal/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Logs server error via RabbitMQ and returns a gRPC error
func (app *application) ServerErrorResponse(message string) error {
	err := app.logViaRabbit("error", message, "log.ERROR")
	if err != nil {
		log.Printf("error logging via rabbitmq: %v", err.Error())
	}
	return status.Errorf(codes.Internal, "the server encountered a problem and could not process your request")
}

// Logs server error via RabbitMQ and returns a gRPC error
func (app *application) EditConflictResponse(message string) error {
	err := app.logViaRabbit("error", message, "log.ERROR")
	if err != nil {
		log.Printf("error logging via rabbitmq: %v", err.Error())
	}
	return status.Errorf(codes.Internal, "unable to update the record due to an edit conflict, please try again")
}

// Validates user input data and returns an gRPC error if validation fails
func (app *application) checkValidationStatus(v *validator.Validator) error {
	if !v.Valid() {
		var errorMessages []string
		for _, errMsg := range v.Errors {
			errorMessages = append(errorMessages, errMsg)
		}

		combinedErrors := strings.Join(errorMessages, ", ")
		return status.Errorf(codes.InvalidArgument, "Errors: %s", combinedErrors)
	}

	return nil
}
