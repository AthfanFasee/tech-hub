package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.RedirectFixedPath = true
	router.RedirectTrailingSlash = true

	// Converting our err helpers as handlers and using them instead of default err handlers
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Application routes
	// router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthCheckHandler)
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())
	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthCheckHandler)

	// Post routes
	router.HandlerFunc(http.MethodGet, "/api/v1/posts", app.showPostsHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/post/:id", app.showSinglePostHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/post", app.requireActivatedUser(app.createPostHandler))
	router.HandlerFunc(http.MethodPatch, "/api/v1/post/:id", app.requireActivatedUser(app.updatePostHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/post/:id", app.requireActivatedUser(app.deletePostHandler))
	router.HandlerFunc(http.MethodPatch, "/api/v1/posts/like/:id", app.requireActivatedUser(app.likePostHandler))
	router.HandlerFunc(http.MethodPatch, "/api/v1/posts/dislike/:id", app.requireActivatedUser(app.dislikePostHandler))

	// Comment routes
	router.HandlerFunc(http.MethodGet, "/api/v1/posts/comments/:id", app.showCommentsForPostHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/posts/comment", app.requireActivatedUser(app.createCommentHandler))
	router.HandlerFunc(http.MethodDelete, "/api/v1/posts/comment/:id", app.requireActivatedUser(app.deleteCommentHandler))

	// Authentication routes
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/register", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/api/v1/auth/activate", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/login", app.authenticateUserHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/delete", app.requireActivatedUser(app.deleteUserHandler))

	// Payment routes
	router.HandlerFunc(http.MethodPost, "/api/v1/payment/payment-intent", app.getPaymentIntent)
	router.HandlerFunc(http.MethodPost, "/api/v1/payment/subscribe", app.createCustomerAndSubscribeToPlan)
	router.HandlerFunc(http.MethodPost, "/api/v1/payment/refund", app.refundCharge)
	router.HandlerFunc(http.MethodPost, "/api/v1/payment/cancel-subscription", app.cancelSubscription)

	return app.recoverPanic(app.secureHeaders(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
