package main

import (
	"io"
	"net/http"

	"github.com/AthfanFasee/broker/internal/validator"
	"github.com/AthfanFasee/broker/proto/posts"
	"google.golang.org/protobuf/encoding/protojson"
)

// Retrieves and returns posts based on query parameters
func (app *application) showPostsHandler(w http.ResponseWriter, r *http.Request) {
	var getPostsRequestData *posts.Empty

	// Get the url.Values map containing the query string data
	queryString := r.URL.Query()

	v := validator.New()

	sort := app.readString(queryString, "sort", "-id")
	title := app.readString(queryString, "title", "")
	page := app.readInt(queryString, "page", 1, v)
	limit := app.readInt(queryString, "limit", 6, v)
	userID := int64(app.readInt(queryString, "id", 0, v))

	result := app.GetPosts(w, r, getPostsRequestData, sort, title, page, userID, limit)

	err := app.writeJSON(w, http.StatusOK, envelope{"posts": result.Posts, "metadata": result.MetaData}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Retrieves and returns a single post
func (app *application) showSinglePostHandler(w http.ResponseWriter, r *http.Request) {
	var getPostRequestData *posts.GetPostRequest

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	getPostRequestData.Id = id

	result := app.GetPost(w, r, getPostRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"post": result.Post}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Creates and returns the created post
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {

	var creatPostRequestData *posts.CreatePostRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = protojson.Unmarshal(body, creatPostRequestData)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Get user info from request context
	userID, ok := r.Context().Value("user-id").(int64)
	if !ok {
		app.serverErrorResponse(w, r, err)
		return
	}
	userName, ok := r.Context().Value("user-name").(string)
	if !ok {
		app.serverErrorResponse(w, r, err)
		return
	}

	creatPostRequestData.UserId = userID
	creatPostRequestData.UserName = userName

	result := app.CreatePost(w, r, creatPostRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"post": result.Post}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Updates and returns the updated post
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	var updatePostRequestData *posts.UpdatePostRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = protojson.Unmarshal(body, updatePostRequestData)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	result := app.UpdatePost(w, r, updatePostRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"post": result.Post}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// Deletes a post
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	var getPostRequestData *posts.GetPostRequest

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	getPostRequestData.Id = id

	result := app.DeletePost(w, r, getPostRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"message": result.Message}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Likes a post
func (app *application) likePostHandler(w http.ResponseWriter, r *http.Request) {
	var likePostRequestData *posts.LikePostRequest

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	likePostRequestData.Id = id

	result := app.LikePost(w, r, likePostRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"post": result.Post}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Dislikes a post
func (app *application) dislikePostHandler(w http.ResponseWriter, r *http.Request) {
	var likePostRequestData *posts.LikePostRequest

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	likePostRequestData.Id = id

	result := app.DislikePost(w, r, likePostRequestData)

	err = app.writeJSON(w, http.StatusOK, envelope{"post": result.Post}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
