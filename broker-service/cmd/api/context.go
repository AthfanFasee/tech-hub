package main

import (
	"context"
	"net/http"
)

// Prevent naming collisions in request context by defining a custom type
type contextKey string

const userIdContextKey = contextKey("user-id")
const userNameContextKey = contextKey("user-name")
const userActivatedContextKey = contextKey("user-activated")
const userAuthenticatedContextKey = contextKey("user-authenticated")

func (app *application) contextSetUserInfo(r *http.Request, userID int64, userName string, userActivated bool, userAuthenticated bool) *http.Request {
	ctx := context.WithValue(r.Context(), userIdContextKey, userID)
	ctx = context.WithValue(ctx, userNameContextKey, userName)
	ctx = context.WithValue(ctx, userActivatedContextKey, userActivated)
	ctx = context.WithValue(ctx, userAuthenticatedContextKey, userAuthenticated)
	return r.WithContext(ctx)
}

func (app *application) contextGetUserInfo(r *http.Request) (int64, string, bool) {
	userID, ok := r.Context().Value(userIdContextKey).(int64)
	if !ok {
		panic("userID value is missing in request context")
	}
	userName, ok := r.Context().Value(userIdContextKey).(string)
	if !ok {
		panic("userID value is missing in request context")
	}
	userActivated, ok := r.Context().Value(userActivatedContextKey).(bool)
	if !ok {
		panic("userID value is missing in request context")
	}

	return userID, userName, userActivated
}

func (app *application) contextGetUserAuthenticatedStatus(r *http.Request) bool {
	userAuthenticated, ok := r.Context().Value(userActivatedContextKey).(bool)
	if !ok {
		panic("userID value is missing in request context")
	}

	return userAuthenticated
}
