package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"runtime/debug"
	"strings"
	"time"

	"github.com/AthfanFasee/authentication/event"
	"github.com/AthfanFasee/authentication/internal/data"
	"github.com/AthfanFasee/authentication/internal/validator"
	users "github.com/AthfanFasee/authentication/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	users.UnimplementedUserServiceServer
	Models      data.Models
	DB          *sql.DB
	Application *application
}

type RabbitPayload struct {
	Name string
	Data any
}

// Gets a single user by id
func (u *UserService) GetUser(ctx context.Context, req *users.GetUserRequest) (*users.GetUserResponse, error) {
	id := req.GetId()
	result, err := u.Models.Users.GetOne(int64(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "requested resource is not found")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	user := users.User{
		Id:        int64(result.ID),
		Email:     result.Email,
		Name:      result.Name,
		Admin:     result.Admin,
		Bio:       result.Bio,
		Avatar:    result.Avatar,
		Active:    result.Activated,
		CreatedAt: timestamppb.New(result.CreatedAt),
	}

	res := &users.GetUserResponse{User: &user}
	return res, nil
}

// Gets a single user by email
func (u *UserService) GetUserByEmail(ctx context.Context, req *users.GetUserByEmailRequest) (*users.GetUserResponse, error) {
	email := req.GetEmail()
	result, err := u.Models.Users.GetByEmail(email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.Unauthenticated, "invalid email or password")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	user := users.User{
		Id:        int64(result.ID),
		Email:     result.Email,
		Name:      result.Name,
		Admin:     result.Admin,
		Bio:       result.Bio,
		Avatar:    result.Avatar,
		Active:    result.Activated,
		CreatedAt: timestamppb.New(result.CreatedAt),
	}

	res := &users.GetUserResponse{User: &user}
	return res, nil
}

// Deletes a single user by id
func (u *UserService) DeleteUser(ctx context.Context, req *users.GetUserRequest) (*users.RegisterResponse, error) {
	id := req.GetId()
	err := u.Models.Users.DeleteByID(int64(id))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.NotFound, "requested resource is not found")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	res := &users.RegisterResponse{Message: "user deleted successfully"}
	return res, nil
}

// Register a user using ther credentials
func (u *UserService) Register(ctx context.Context, req *users.RegisterRequest) (*users.RegisterResponse, error) {
	email := req.GetEmail()
	name := req.GetName()
	password := req.GetPassword()

	user := &data.User{
		Name:      strings.TrimSpace(name),
		Email:     strings.TrimSpace(email),
		Admin:     false,
		Activated: false,
	}

	err := user.Password.Set(password)
	if err != nil {
		err := u.Application.ServerErrorResponse(err.Error())
		if err != nil {
			return nil, err
		}
	}

	v := validator.New()

	data.ValidateUser(v, user)

	err = u.Application.ValidationFailedResponse(v)
	if err != nil {
		return nil, err
	}

	id, err := u.Models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", " email address already exists")
			err = u.Application.ValidationFailedResponse(v)
			if err != nil {
				return nil, err
			}
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	token, err := u.Models.Tokens.New(int64(id), 24*time.Hour, data.ScopeActivation)
	if err != nil {
		err := u.Application.ServerErrorResponse(err.Error())
		if err != nil {
			return nil, err
		}
	}

	// Sends mail to the user via RabbitMQ
	mailData := map[string]interface{}{
		"activationToken": token.Plaintext,
		"userID":          id,
	}
	err = u.Application.pushToQueue("mail", mailData, "mailUser")
	if err != nil {
		log.Printf("error sending mail via rabbitmq : %v", err.Error())
	}

	res := &users.RegisterResponse{Message: "user created successfully"}
	return res, nil
}

// Logs in user, create a token and send it to them.
func (u *UserService) Authenticate(ctx context.Context, req *users.AuthenticateRequest) (*users.AuthenticateResponse, error) {
	email := req.GetEmail()
	password := req.GetPassword()

	v := validator.New()
	data.ValidateEmail(v, email)
	data.ValidatePasswordPlaintext(v, password)

	err := u.Application.checkValidationStatus(v)
	if err != nil {
		return nil, err
	}

	user, err := u.Application.Models.Users.GetByEmail(email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return nil, status.Errorf(codes.Unauthenticated, "invalid email or password")
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	match, err := user.Password.PasswordMatches(password)
	if err != nil {
		err := u.Application.ServerErrorResponse(err.Error())
		if err != nil {
			return nil, err
		}
	}

	if !match {
		return nil, status.Errorf(codes.Unauthenticated, "invalid email or password")
	}

	if !user.Activated {
		return nil, status.Errorf(codes.Unauthenticated, "your user account must be activated")
	}

	token, err := u.Application.Models.Tokens.New(int64(user.ID), 30*24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		err := u.Application.ServerErrorResponse(err.Error())
		if err != nil {
			return nil, err
		}
	}

	tokenResult := &users.Token{
		Plaintext: token.Plaintext,
		Hash:      token.Hash,
		UserId:    token.UserID,
		Expiry:    timestamppb.New(token.Expiry),
		Scope:     token.Scope,
	}
	res := &users.AuthenticateResponse{Token: tokenResult, UserName: user.Name}
	return res, nil
}

// Changes a user's status to activated
func (u *UserService) Activate(ctx context.Context, req *users.ActivateUserRequest) (*users.RegisterResponse, error) {
	tokenPlainText := req.GetTokenPlainText()

	v := validator.New()
	data.ValidateTokenPlainText(v, tokenPlainText)
	err := u.Application.ValidationFailedResponse(v)
	if err != nil {
		return nil, err
	}

	user, err := u.Application.Models.Users.GetForToken(data.ScopeActivation, tokenPlainText)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired token")
			err = u.Application.ValidationFailedResponse(v)
			if err != nil {
				return nil, err
			}
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	user.Activated = true

	err = u.Application.Models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			err := u.Application.EditConflictResponse(err.Error())
			if err != nil {
				return nil, err
			}
		default:
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}
	}

	err = u.Application.Models.Tokens.DeleteAllForUser(data.ScopeActivation, int64(user.ID))
	if err != nil {
		err := u.Application.ServerErrorResponse(err.Error())
		if err != nil {
			return nil, err
		}
	}

	res := &users.RegisterResponse{Message: "user activated successfully"}
	return res, nil
}

// Pushes an event to RabbitMQ
func (app *application) pushToQueue(name string, data any, key string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := RabbitPayload{
		Name: name,
		Data: data,
	}

	j, _ := json.MarshalIndent(&payload, "", "\t")
	err = emitter.Push(string(j), key)
	if err != nil {
		return err
	}

	return nil
}

// Logs error via RabbitMQ
func (app *application) logViaRabbit(name, errorMessage, severity string) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	stackTrace := string(debug.Stack())

	logMessage := fmt.Sprintf("Timestamp: %s\nError: %s\nStackTrace:\n%s", timestamp, errorMessage, stackTrace)

	err := app.pushToQueue(name, logMessage, severity)
	if err != nil {
		return err
	}

	return nil
}

// Starts listening to gRPC calls
func (app *application) gRPCListen() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", app.config.gRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	s := grpc.NewServer()
	users.RegisterUserServiceServer(s, &UserService{Models: app.Models, DB: app.DB, Application: app})

	// Health Check
	dbSystem := "database" // Specify the system as "database"
	healthcheck := health.NewServer()
	healthpb.RegisterHealthServer(s, healthcheck)

	go func() {
		// Asynchronously inspect database status and toggle serving status as needed
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			// Perform a health check on the database
			err := app.DB.PingContext(ctx)
			cancel()

			// Set the serving status based on the database health check result
			if err != nil {
				healthcheck.SetServingStatus(dbSystem, healthpb.HealthCheckResponse_NOT_SERVING)
			} else {
				healthcheck.SetServingStatus(dbSystem, healthpb.HealthCheckResponse_SERVING)
			}

			// Change this in future, so that check happens only when errors occur
			time.Sleep(time.Second * 60)
		}
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	log.Printf("gRPC Server started on port %v", app.config.gRPCPort)
}
