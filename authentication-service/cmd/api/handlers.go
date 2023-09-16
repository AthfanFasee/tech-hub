package main

import (
	"encoding/json"
	"net/http"
)

type jsonRes struct {
	Message string `json:"message"`
}

func (app *application) Authenticate(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	payload := jsonRes{
		Message: "User loggedIn",
	}

	out, _ := json.MarshalIndent(payload, "", "\t")
	w.Write(out)
}
