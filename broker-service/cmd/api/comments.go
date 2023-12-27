package main

import (
	"io"
	"net/http"

	"github.com/AthfanFasee/broker/proto/posts"
	"google.golang.org/protobuf/encoding/protojson"
)

// Retrieves and returns comments of a single post
func (app *application) showCommentsForPostHandler(w http.ResponseWriter, r *http.Request) {
	var getPostRequestData *posts.GetPostRequest

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	getPostRequestData.Id = id

	result := app.GetCommentsForPost(w, r, getPostRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"comments": result.Comments}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Creates and returns the created comment
func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	var creatCommentRequestData *posts.CreateCommentRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = protojson.Unmarshal(body, creatCommentRequestData)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Add values from req.CONTEXT

	result := app.CreateComment(w, r, creatCommentRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"comment": result.Comment}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Deletes a comment
func (app *application) deleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	var getPostRequestData *posts.GetPostRequest

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	getPostRequestData.Id = id

	result := app.DeleteComment(w, r, getPostRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"message": result.Message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
