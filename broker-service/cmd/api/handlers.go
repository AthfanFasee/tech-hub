package main

import (
	"encoding/json"
	"net/http"
)

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
