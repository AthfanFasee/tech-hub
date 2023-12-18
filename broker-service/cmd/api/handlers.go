package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AthfanFasee/broker/event"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type RabbitPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type jsonRes struct {
	Message string `json:"message"`
}

func (app *application) broker(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)

	app.writeJSON(w, http.StatusAccepted, envelope{"message": "Broker service"}, nil)

}

func (app *application) handleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, r, requestPayload.Auth)

	default:
		app.badRequestResponse(w, r, errors.New("unknown action"))
	}

}

// func (app *application) logItem(w http.ResponseWriter, r *http.Request, entry LogPayload) {
// 	jsonData, _ := json.MarshalIndent(entry, "", "\t")

// 	logServiceURL := "http://logger-service/log"

// 	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		app.badRequestResponse(w, r, err)
// 		return
// 	}

// 	request.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}

// 	response, err := client.Do(request)
// 	if err != nil {
// 		app.badRequestResponse(w, r, err)
// 		return
// 	}

// 	defer response.Body.Close()

// 	if response.StatusCode != http.StatusAccepted {
// 		app.badRequestResponse(w, r, err)
// 		return
// 	}

// 	app.writeJSON(w, http.StatusAccepted, envelope{"message": "logged"}, nil)
// }

func (app *application) authenticate(w http.ResponseWriter, r *http.Request, a AuthPayload) {
	// Create some json and send to auth service
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	// call service
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	defer response.Body.Close()

	// make sure we get abck correct status code

	if response.StatusCode == http.StatusUnauthorized {
		app.invalidCredentialsResponse(w, r)
		return
	} else if response.StatusCode != http.StatusAlreadyReported {
		app.serverErrorResponse(w, r, errors.New("error calling auth service"))
		return
	}

	// create a var we'll read response.body into

	var jsonFromService jsonRes

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	app.writeJSON(w, http.StatusAccepted, envelope{"message": jsonFromService.Message}, nil)
}

// func (app *application) logEventViaRabbitMQ(w http.ResponseWriter, r *http.Request, l LogPayload) {
// 	err := app.pushToQueue(l.Name, l.Data)
// 	if err != nil {
// 		app.badRequestResponse(w, r, err)
// 		return
// 	}

// 	app.writeJSON(w, http.StatusAccepted, envelope{"message": "logged via rabbit"}, nil)
// }

func (app *application) pushToQueue(name, message string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := RabbitPayload{
		Name: name,
		Data: message,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}
