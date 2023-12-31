package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/AthfanFasee/broker/proto/posts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Makes a gRPC call to retrieve posts from the posts service
func (app *application) GetPosts(w http.ResponseWriter, r *http.Request, getPostsRequestData *posts.Empty, sort string, title string, page int, userID int64, limit int) *posts.GetPostsResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return &posts.GetPostsResponse{}
	}

	defer conn.Close()

	c := posts.NewPostsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// Send query data as gRPC meta data
	header := metadata.New(map[string]string{
		"sort":    sort,
		"title":   title,
		"user-id": strconv.FormatInt(userID, 10),
		"page":    strconv.Itoa(page),
		"limit":   strconv.Itoa(limit),
	})
	grpcCtx := metadata.NewOutgoingContext(ctx, header)

	result, err := c.GetPosts(grpcCtx, getPostsRequestData)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			app.notFoundResponse(w, r)
			return &posts.GetPostsResponse{}
		}

		app.serverErrorResponse(w, r, err)
		return &posts.GetPostsResponse{}
	}

	return result
}

// Makes a gRPC call to retrieve a single post from the posts service
func (app *application) GetPost(w http.ResponseWriter, r *http.Request, getPostRequestData *posts.GetPostRequest) *posts.GetPostResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return &posts.GetPostResponse{}
	}

	defer conn.Close()

	c := posts.NewPostsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.GetPost(ctx, getPostRequestData)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			app.notFoundResponse(w, r)
			return &posts.GetPostResponse{}
		}

		app.serverErrorResponse(w, r, err)
		return &posts.GetPostResponse{}
	}

	return result
}

// Makes a gRPC call to create a post in posts service
func (app *application) CreatePost(w http.ResponseWriter, r *http.Request, createPostRequestRequestData *posts.CreatePostRequest) *posts.GetPostResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return &posts.GetPostResponse{}
	}

	defer conn.Close()

	c := posts.NewPostsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.CreatePost(ctx, createPostRequestRequestData)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.InvalidArgument {
			app.validationFailedResponse(w, r, err.Error())
			return &posts.GetPostResponse{}
		}

		app.serverErrorResponse(w, r, err)
		return &posts.GetPostResponse{}
	}

	return result
}

// Makes a gRPC call to update a post in posts service
func (app *application) UpdatePost(w http.ResponseWriter, r *http.Request, updatePostRequestRequestData *posts.UpdatePostRequest) *posts.GetPostResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return &posts.GetPostResponse{}
	}

	defer conn.Close()

	c := posts.NewPostsServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.UpdatePost(ctx, updatePostRequestRequestData)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			app.notFoundResponse(w, r)
			return &posts.GetPostResponse{}
		} else if ok && st.Code() == codes.InvalidArgument {
			app.validationFailedResponse(w, r, err.Error())
			return &posts.GetPostResponse{}
		} else if ok && st.Code() == codes.AlreadyExists {
			app.editConflictResponse(w, r)
			return &posts.GetPostResponse{}
		}

		app.serverErrorResponse(w, r, err)
		return &posts.GetPostResponse{}
	}

	return result
}

// Makes a gRPC call to delete a post in posts service
func (app *application) DeletePost(w http.ResponseWriter, r *http.Request, getPostRequestData *posts.GetPostRequest) *posts.DeletePostResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return &posts.DeletePostResponse{}
	}

	defer conn.Close()

	c := posts.NewPostsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.DeletePost(ctx, getPostRequestData)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			app.notFoundResponse(w, r)
			return &posts.DeletePostResponse{}
		}

		app.serverErrorResponse(w, r, err)
		return &posts.DeletePostResponse{}
	}

	return result
}

// Makes a gRPC call to like a post in posts service
func (app *application) LikePost(w http.ResponseWriter, r *http.Request, likePostRequestData *posts.LikePostRequest) *posts.GetPostResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return &posts.GetPostResponse{}
	}

	defer conn.Close()

	c := posts.NewPostsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.LikePost(ctx, likePostRequestData)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			app.notFoundResponse(w, r)
			return &posts.GetPostResponse{}
		} else if ok && st.Code() == codes.InvalidArgument {
			app.badRequestResponse(w, r, err)
			return &posts.GetPostResponse{}
		}

		app.serverErrorResponse(w, r, err)
		return &posts.GetPostResponse{}
	}

	return result
}

// Makes a gRPC call to dislike a post in posts service
func (app *application) DislikePost(w http.ResponseWriter, r *http.Request, likePostRequestData *posts.LikePostRequest) *posts.GetPostResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return &posts.GetPostResponse{}
	}

	defer conn.Close()

	c := posts.NewPostsServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.DislikePost(ctx, likePostRequestData)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			app.notFoundResponse(w, r)
			return &posts.GetPostResponse{}
		} else if ok && st.Code() == codes.InvalidArgument {
			app.badRequestResponse(w, r, err)
			return &posts.GetPostResponse{}
		}

		app.serverErrorResponse(w, r, err)
		return &posts.GetPostResponse{}
	}

	return result
}
