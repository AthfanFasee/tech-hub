package main

import (
	"context"
	"fmt"
	"log"
	"net"

	mail "github.com/AthfanFasee/mail/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MailService struct {
	mail.UnimplementedMailServiceServer
	Application *application
}

// Register a user using their credentials
func (u *MailService) SendWelcomeMail(ctx context.Context, req *mail.SendWelcomeMailRequest) (*mail.SendWelcomeMailResponse, error) {
	from := req.GetFrom()
	to := req.GetTo()
	subject := req.GetSubject()
	message := req.GetMessage()

	msg := Message{
		From:    from,
		To:      to,
		Subject: subject,
		Data:    message,
	}

	err := u.Application.Mailer.SendSMTPMessage(msg)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("bad request: %v", err.Error()))
	}

	res := &mail.SendWelcomeMailResponse{Message: "sent mail to " + to}
	return res, nil
}

// Starts listening to gRPC calls
func (app *application) gRPCListen() {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", app.config.gRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	s := grpc.NewServer()
	mail.RegisterMailServiceServer(s, &MailService{Application: app})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	log.Printf("gRPC Server started on port %v", app.config.gRPCPort)
}
