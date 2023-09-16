package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func (app *application) Authenticate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)

	err := app.logRequest("authentication", "A user is logged in")
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	app.writeJSON(w, http.StatusAccepted, envelope{"message": "logged In"}, nil)
}

func (app *application) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}

	_, err = client.Do(request)

	if err != nil {
		return err
	}
	return nil
}
