package main

import (
	"context"
	"net/http"
	"time"

	users "github.com/AthfanFasee/authentication/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Makes a gRPC call to register a user in authentication service
func (app *application) Register(w http.ResponseWriter, r *http.Request, registerRequestData *users.RegisterRequest) *users.RegisterResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	defer conn.Close()

	c := users.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.Register(ctx, registerRequestData)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.InvalidArgument {
			app.validationFailedResponse(w, r, err.Error())
		}
		app.serverErrorResponse(w, r, err)
	}

	return result
}

// Makes a gRPC call to authenticate a user in authentication service
func (app *application) Authenticate(w http.ResponseWriter, r *http.Request, authenticationRequestData *users.AuthenticateRequest) *users.AuthenticateResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	defer conn.Close()

	c := users.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.Authenticate(ctx, authenticationRequestData)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			app.invalidCredentialsResponse(w, r)
		} else if ok && st.Code() == codes.InvalidArgument {
			app.validationFailedResponse(w, r, err.Error())
		}
		app.serverErrorResponse(w, r, err)
	}

	return result
}

// Makes a gRPC call to activate a user in authentication service
func (app *application) Activate(w http.ResponseWriter, r *http.Request, activateRequestData *users.ActivateRequest) *users.RegisterResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	defer conn.Close()

	c := users.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.Activate(ctx, activateRequestData)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.InvalidArgument {
			app.validationFailedResponse(w, r, err.Error())
		}
		app.serverErrorResponse(w, r, err)
	}
	return result
}

// Makes a gRPC call to delete a user in authentication service
func (app *application) DeleteUser(w http.ResponseWriter, r *http.Request, getUserRequestData *users.GetUserRequest) *users.RegisterResponse {
	conn, err := grpc.Dial("posts-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	defer conn.Close()

	c := users.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := c.DeleteUser(ctx, getUserRequestData)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			app.notFoundResponse(w, r)
		}
		app.serverErrorResponse(w, r, err)
	}

	// BEFORE RETURNING PUSH event to delete this user's posts
	return result
}
