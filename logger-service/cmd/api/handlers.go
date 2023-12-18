package main

import (
	"net/http"

	"github.com/AthfanFasee/logger-service/data"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *application) WriteLog(w http.ResponseWriter, r *http.Request) {
	var requestPayload JSONPayload

	_ = app.readJSON(w, r, &requestPayload)

	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	err := app.Models.LogEntry.Insert(event)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, envelope{"message": "comment deleted successfully"}, nil)
}
