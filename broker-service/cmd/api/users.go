package main

import (
	"io"
	"net/http"

	users "github.com/AthfanFasee/broker/proto/users"
	"google.golang.org/protobuf/encoding/protojson"
)

// Registers a user
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var registerRequestData *users.RegisterRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = protojson.Unmarshal(body, registerRequestData)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	result := app.Register(w, r, registerRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"message": result.Message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Authenticates a user
func (app *application) authenticateUserHandler(w http.ResponseWriter, r *http.Request) {
	var authenticateRequestData *users.AuthenticateRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = protojson.Unmarshal(body, authenticateRequestData)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	result := app.Authenticate(w, r, authenticateRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"authentication_token": result.Token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Activates a user
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var activateRequestData *users.ActivateRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = protojson.Unmarshal(body, activateRequestData)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	result := app.Activate(w, r, activateRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"message": result.Message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Deletes a user
func (app *application) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var getUserRequestData *users.GetUserRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = protojson.Unmarshal(body, getUserRequestData)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	result := app.DeleteUser(w, r, getUserRequestData)

	// Before sending a response push an event to rabbitMQ to delete this user's posts
	app.pushToQueue("deletePost", getUserRequestData.Id, "posts")

	err = app.writeJSON(w, http.StatusOK, envelope{"message": result.Message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
