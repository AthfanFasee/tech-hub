package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/AthfanFasee/logger-service/data"
	logs "github.com/AthfanFasee/logger-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

// CHANGE THE WAY ERRORS R HANDLED

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetLogEntry()

	// Write the log
	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		log.Printf("failed to insert log entry: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to insert log entry")
	}

	// Return response
	res := &logs.LogResponse{Message: "succesfully logged to mongoDB"}
	return res, err

}

func (app *application) gRPCListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", app.config.gRPCPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	s := grpc.NewServer()

	logs.RegisterLogServiceServer(s, &LogServer{Models: app.Models})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	log.Printf("gRPC Server started on port %v", app.config.gRPCPort)
}
