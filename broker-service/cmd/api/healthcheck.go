package main

import "net/http"

// Gets server's availablity status
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	serverData := envelope{
		"status": "available",
	}

	err := app.writeJSON(w, http.StatusOK, serverData, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
