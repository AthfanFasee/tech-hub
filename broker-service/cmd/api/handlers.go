package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type jsonRes struct {
	Message string `json:"message"`
}

func (app *application) broker(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	payload := jsonRes{
		Message: "Broker Server",
	}

	out, _ := json.MarshalIndent(payload, "", "\t")
	w.Write(out)
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

	var payload jsonRes
	payload.Message = "Authenticated!"

	app.writeJSON(w, http.StatusAccepted, make(envelope), nil)
}
