package main

import (
	"context"
	"net/http"
	"time"

	"github.com/AthfanFasee/broker/proto/posts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Makes a gRPC call to retrieve comments of a single post from the posts service
func (app *application) GetCommentsForPost(w http.ResponseWriter, r *http.Request, getPostRequestData *posts.GetPostRequest) *posts.GetCommentsForPostResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	defer conn.Close()

	c := posts.NewPostsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.GetCommentsForPost(ctx, getPostRequestData)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			app.notFoundResponse(w, r)
		}
		app.serverErrorResponse(w, r, err)
	}

	return result
}

// Makes a gRPC call to create a comment in posts service
func (app *application) CreateComment(w http.ResponseWriter, r *http.Request, createCommentRequestRequestData *posts.CreateCommentRequest) *posts.CreateCommentResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	defer conn.Close()

	c := posts.NewPostsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.CreateComment(ctx, createCommentRequestRequestData)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.InvalidArgument {
			app.validationFailedResponse(w, r, err.Error())
		}
		app.serverErrorResponse(w, r, err)
	}

	return result
}

// Makes a gRPC call to delete a comment in posts service
func (app *application) DeleteComment(w http.ResponseWriter, r *http.Request, getPostRequestData *posts.GetPostRequest) *posts.DeletePostResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	defer conn.Close()

	c := posts.NewPostsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.DeleteComment(ctx, getPostRequestData)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			app.notFoundResponse(w, r)
		}
		app.serverErrorResponse(w, r, err)
	}

	return result
}
