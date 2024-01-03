package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type stripePayload struct {
	Currency      string `json:"currency"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Email         string `json:"email"`
	CardBrand     string `json:"card_brand"`
	ExpiryMonth   int    `json:"exp_month"`
	ExpiryYear    int    `json:"exp_year"`
	LastFour      string `json:"last_four"`
	IsRecurring   bool   `json:"is_recurring"`
	Plan          string `json:"plan"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
}

type chargeToRefund struct {
	Id            int    `json:"id"`
	PaymentIntent string `json:"pi"`
	Amount        int    `json:"amount"`
	Currency      string `json:"currency"`
}

type subToCancel struct {
	ID            int    `json:"id"`
	PaymentIntent string `json:"pi"`
	Currency      string `json:"currency"`
}

type getPaymentIntentRespone struct {
	Pi      any
	Message string
}

// Makes http request to payment-service to get payment intent
func (app *application) getPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var stripePayload stripePayload

	// Decoding JSON values in to input struct
	err := app.readJSON(w, r, &stripePayload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	jsonData, err := json.MarshalIndent(stripePayload, "", "\t")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// call the service
	request, err := http.NewRequest("POST", "http://payment-service/payment-intent", bytes.NewBuffer(jsonData))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode != http.StatusAccepted {
		app.serverErrorResponse(w, r, err)
		return
	}

	var getPaymentIntentRespone getPaymentIntentRespone

	err = json.NewDecoder(response.Body).Decode(response)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"payment_intent": getPaymentIntentRespone.Pi, "message": getPaymentIntentRespone.Message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Makes http request to payment-service to create a customer and subscribe to plan
func (app *application) createCustomerAndSubscribeToPlan(w http.ResponseWriter, r *http.Request) {
	var stripePayload stripePayload

	// Decoding JSON values in to input struct
	err := app.readJSON(w, r, &stripePayload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	jsonData, err := json.MarshalIndent(stripePayload, "", "\t")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// call the service
	request, err := http.NewRequest("POST", "http://payment-service/subscribe-to-plan", bytes.NewBuffer(jsonData))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode != http.StatusAccepted {
		app.serverErrorResponse(w, r, err)
		return
	}

	var message string

	err = json.NewDecoder(response.Body).Decode(response)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Makes http request to payment-service to refund charge
func (app *application) refundCharge(w http.ResponseWriter, r *http.Request) {
	var chargeToRefund chargeToRefund

	// Decoding JSON values in to input struct
	err := app.readJSON(w, r, &chargeToRefund)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	jsonData, err := json.MarshalIndent(chargeToRefund, "", "\t")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// call the service
	request, err := http.NewRequest("POST", "http://payment-service/refund", bytes.NewBuffer(jsonData))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode != http.StatusAccepted {
		app.serverErrorResponse(w, r, err)
		return
	}

	var message string

	err = json.NewDecoder(response.Body).Decode(response)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Makes http request to payment-service to cancel subscription
func (app *application) cancelSubscription(w http.ResponseWriter, r *http.Request) {
	var subToCancel subToCancel

	// Decoding JSON values in to input struct
	err := app.readJSON(w, r, &subToCancel)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	jsonData, err := json.MarshalIndent(subToCancel, "", "\t")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// call the service
	request, err := http.NewRequest("POST", "http://cancel-subscription", bytes.NewBuffer(jsonData))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer response.Body.Close()

	// make sure we get back the correct status code
	if response.StatusCode != http.StatusAccepted {
		app.serverErrorResponse(w, r, err)
		return
	}

	var message string

	err = json.NewDecoder(response.Body).Decode(response)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
