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

// Logs server error to stdout and sends server error response
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}
