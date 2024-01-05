package main

import (
	"fmt"
	"log"
	"net/http"
)

// Logs error to stdout
func (app *application) logError(r *http.Request, err error) {
	logMessage := fmt.Sprintf("Error: %v, request_method: %s, request_url: %s",
		err, r.Method, r.URL.String())
	log.Println(logMessage)
}

// Sends error response to client
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	// Incase of not able to write a JSON err response, log the err and send the user a 500 err code by default
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// Logs server error to stdout and via RabbitMQ
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	app.logViaRabbit("error", err.Error(), "log.ERROR")

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// Sends bad request response
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// Sends not found response
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "requested resource is not found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// Sends method not allowed response
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("method %s is not supported", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// Sends validation failed response
func (app *application) validationFailedResponse(w http.ResponseWriter, r *http.Request, errors string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// Sends edit conflict response
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

// Sends too many requests response
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

// Sends unauthorized response
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// Sends unauthorized response incase of invalid credentials
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid email or password"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// Sends unauthorized response incase of invalid token
func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	// Inform client that they are expected to authenticate using a bearer token
	w.Header().Set("WWWW-Authenticate", "Bearer")

	message := "authentication token is invalid or missing"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// Sends forbidden response
func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated"
	app.errorResponse(w, r, http.StatusForbidden, message)
}
