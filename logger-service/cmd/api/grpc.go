package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/AthfanFasee/logger-service/data"
	logs "github.com/AthfanFasee/logger-service/proto"
	"google.golang.org/grpc"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

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
		res := &logs.LogResponse{Error: true, Message: fmt.Sprintf("failed to insert log entry: %v", err)}
		return res, err
	}

	// Return response
	res := &logs.LogResponse{Error: false, Message: "succesfully logged to mongoDB"}
	return res, err

}

func (app *application) gRPCListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", gRpcPORT))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	s := grpc.NewServer()

	logs.RegisterLogServiceServer(s, &LogServer{Models: app.Models})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	log.Printf("gRPC Server started on port %s", gRpcPORT)
}
