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

	// Post routes
	router.HandlerFunc(http.MethodGet, "/api/v1/posts", app.showPostsHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/post/:id", app.showSinglePostHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/post", app.authenticate(http.HandlerFunc(app.createPostHandler)))
	router.HandlerFunc(http.MethodPatch, "/api/v1/post/:id", app.authenticate(http.HandlerFunc(app.updatePostHandler)))
	router.HandlerFunc(http.MethodDelete, "/api/v1/post/:id", app.authenticate(http.HandlerFunc(app.deletePostHandler)))
	router.HandlerFunc(http.MethodPatch, "/api/v1/posts/like/:id", app.authenticate(http.HandlerFunc(app.likePostHandler)))
	router.HandlerFunc(http.MethodPatch, "/api/v1/posts/dislike/:id", app.authenticate(http.HandlerFunc(app.dislikePostHandler)))

	// Comment routes
	router.HandlerFunc(http.MethodGet, "/api/v1/posts/comments/:id", app.showCommentsForPostHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/posts/comment", app.authenticate(http.HandlerFunc(app.createCommentHandler)))
	router.HandlerFunc(http.MethodDelete, "/api/v1/posts/comment/:id", app.authenticate(http.HandlerFunc(app.deleteCommentHandler)))

	// Authentication routes
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/register", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/api/v1/auth/activate", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/login", app.authenticateUserHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/auth/delete", app.authenticate(http.HandlerFunc(app.deleteUserHandler)))

	return app.recoverPanic(app.secureHeaders(app.enableCORS(app.rateLimit(router))))
}
