package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.Heartbeat("/ping"))

	router.Post("/api/payment-intent", app.getPaymentIntent)
	router.Post("/api/subscribe-to-plan", app.createCustomerAndSubscribeToPlan)
	router.Post("/api/refund", app.RefundCharge)
	router.Post("/api/cancel-subscription", app.CancelSubscription)

	return router
}
