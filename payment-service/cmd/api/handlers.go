package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AthfanFasee/payment/internal/cards"
	"github.com/AthfanFasee/payment/internal/data"
	"github.com/stripe/stripe-go/v72"
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

// Gets payment intent
func (app *application) getPaymentIntent(w http.ResponseWriter, r *http.Request) {
	var payload stripePayload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	amount, err := strconv.Atoi(payload.Amount)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: payload.Currency,
	}

	pi, msg, err := card.Charge(payload.Currency, amount)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"payment_intent": pi, "message": msg}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// Creates a customer and subscribe to plan
func (app *application) createCustomerAndSubscribeToPlan(w http.ResponseWriter, r *http.Request) {
	var stripeData stripePayload
	err := json.NewDecoder(r.Body).Decode(&stripeData)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.infoLog.Println(stripeData.Email, stripeData.LastFour, stripeData.PaymentMethod, stripeData.Plan)

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: stripeData.Currency,
	}

	okay := true
	var subscription *stripe.Subscription
	txnMsg := "Transaction successful"

	stripeCustomer, msg, err := card.CreateCustomer(stripeData.PaymentMethod, stripeData.Email)
	if err != nil {
		app.logError(r, err)
		okay = false
		txnMsg = msg
	}

	if okay {
		subscription, err = card.SubscribeToPlan(stripeCustomer, stripeData.Plan, stripeData.Email, stripeData.LastFour, "")
		if err != nil {
			app.logError(r, err)
			okay = false
			txnMsg = "Error subscribing customer"
		}
		app.infoLog.Println("subscription id is", subscription.ID)
	}

	if okay {
		customerID, err := app.SaveCustomer(stripeData.FirstName, stripeData.LastName, stripeData.Email)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Create transaction
		amount, err := strconv.Atoi(stripeData.Amount)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		txn := data.Transaction{
			Amount:              amount,
			Currency:            stripeData.Currency,
			LastFour:            stripeData.LastFour,
			ExpiryMonth:         stripeData.ExpiryMonth,
			ExpiryYear:          stripeData.ExpiryYear,
			TransactionStatusID: 2,
			PaymentIntent:       subscription.ID,
			PaymentMethod:       stripeData.PaymentMethod,
		}

		txnID, err := app.SaveTransaction(txn)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Create order
		order := data.Order{
			TransactionID: txnID,
			CustomerID:    customerID,
			IsRecurring:   stripeData.IsRecurring,
			StatusID:      1,
			Amount:        amount,
		}

		_, err = app.SaveOrder(order)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": txnMsg}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

// Saves a customer and returns id
func (app *application) SaveCustomer(firstName, lastName, email string) (int, error) {
	customer := data.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}

	id, err := app.Models.Payment.InsertCustomer(customer)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Saves a transaction and returns id
func (app *application) SaveTransaction(txn data.Transaction) (int, error) {
	id, err := app.Models.Payment.InsertTransaction(txn)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Saves an order and returns id
func (app *application) SaveOrder(order data.Order) (int, error) {
	id, err := app.Models.Payment.InsertOrder(order)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Refunds a charge
func (app *application) RefundCharge(w http.ResponseWriter, r *http.Request) {
	var chargeToRefund struct {
		Id            int    `json:"id"`
		PaymentIntent string `json:"pi"`
		Amount        int    `json:"amount"`
		Currency      string `json:"currency"`
	}

	err := app.readJSON(w, r, &chargeToRefund)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: chargeToRefund.Currency,
	}

	err = card.Refund(chargeToRefund.PaymentIntent, chargeToRefund.Amount)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Update refund status in db
	err = app.Models.Payment.UpdateOrderStatus(chargeToRefund.Id, 2)
	if err != nil {
		app.logViaRabbit("error", err.Error(), "log.ERROR")
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Successful"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// Cancels Subscription
func (app *application) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	var subToCancel struct {
		ID            int    `json:"id"`
		PaymentIntent string `json:"pi"`
		Currency      string `json:"currency"`
	}

	err := app.readJSON(w, r, &subToCancel)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	card := cards.Card{
		Secret:   app.config.stripe.secret,
		Key:      app.config.stripe.key,
		Currency: subToCancel.Currency,
	}

	err = card.CancelSubscription(subToCancel.PaymentIntent)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.Models.Payment.UpdateOrderStatus(subToCancel.ID, 3)
	if err != nil {
		app.logViaRabbit("error", err.Error(), "log.ERROR")
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Successful"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
