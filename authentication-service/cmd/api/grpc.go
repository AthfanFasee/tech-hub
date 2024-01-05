package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AthfanFasee/authentication/internal/data"
	"github.com/AthfanFasee/authentication/internal/validator"
	users "github.com/AthfanFasee/authentication/proto"
	"github.com/dgrijalva/jwt-go"
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
	err := u.Models.Users.Delete(int64(id))
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

	err = u.Application.checkValidationStatus(v)
	if err != nil {
		return nil, err
	}

	id, err := u.Models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", " email address already exists")
			err = u.Application.checkValidationStatus(v)
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

	token, err := u.Application.generateJWTToken(user.ID, user.Name, user.Activated)
	if err != nil {
		err := u.Application.ServerErrorResponse(err.Error())
		if err != nil {
			return nil, err
		}
	}

	// Sends mail to the user via RabbitMQ
	mailData := map[string]interface{}{
		"activationToken": token,
		"userID":          id,
	}
	err = u.Application.pushToQueue("mailUser", mailData, "mail")
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

	token, err := u.Application.generateJWTToken(user.ID, user.Name, user.Activated)
	if err != nil {
		return nil, err
	}

	res := &users.AuthenticateResponse{Token: token}
	return res, nil
}

// Changes a user's status to activated
func (u *UserService) Activate(ctx context.Context, req *users.ActivateRequest) (*users.RegisterResponse, error) {
	tokenString := req.GetTokenPlainText()

	// Load the public key
	keysDir := filepath.Join(".", "keys")
	publicKeyPath := filepath.Join(keysDir, "public.pem")
	publicBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		err := u.Application.ServerErrorResponse(err.Error())
		if err != nil {
			return nil, err
		}
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicBytes)
	if err != nil {
		err := u.Application.ServerErrorResponse(err.Error())
		if err != nil {
			return nil, err
		}
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			err := u.Application.ServerErrorResponse(err.Error())
			if err != nil {
				return nil, err
			}
		}

		// Return the public key for verification
		return publicKey, nil
	})

	var claims jwt.MapClaims
	v := validator.New()

	if claimsMap, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		claims = claimsMap
	} else {
		v.AddError("token", "invalid or expired token")
		err := u.Application.checkValidationStatus(v)
		if err != nil {
			return nil, err
		}
	}

	// Access claims as needed
	userID := claims["user_id"].(int64)

	user, err := u.Application.Models.Users.GetOne(userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired token")
			err = u.Application.checkValidationStatus(v)
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

	res := &users.RegisterResponse{Message: "user activated successfully"}
	return res, nil
}

// Generates JWT token using the private key
func (app *application) generateJWTToken(userID int64, name string, activated bool) (string, error) {
	// Load the private key
	keysDir := filepath.Join(".", "keys")
	privateKeyPath := filepath.Join(keysDir, "private.pem")
	privateBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		err := app.ServerErrorResponse(err.Error())
		if err != nil {
			return "", err
		}
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateBytes)
	if err != nil {
		err := app.ServerErrorResponse(err.Error())
		if err != nil {
			return "", err
		}
	}

	// Create a new token object
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user_id":        userID,
		"user_name":      userID,
		"user_activated": activated,
		"iat":            time.Now(),
		"exp":            time.Now().Add(time.Hour * 24 * 30),
	})

	// Sign and get the complete encoded token as a string using the secret
	token, err := tokenObj.SignedString(privateKey)
	if err != nil {
		err := app.ServerErrorResponse(err.Error())
		if err != nil {
			return "", err
		}
	}

	return token, nil
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

			// TODO : Change this, so that check happens only when errors occur
			time.Sleep(time.Second * 60)
		}
	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	log.Printf("gRPC Server started on port %v", app.config.gRPCPort)
}
