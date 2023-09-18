package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AthfanFasee/log-service/data"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PORT     = "80"
	mongoURL = "mongodb://mongo:27017"
	gRpcPORT = "5001"
)

var client *mongo.Client

type application struct {
	Models data.Models
}

func main() {
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	// Create a context in order to disconnect DB
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := application{
		Models: data.New(client),
	}

	go app.gRPCListen()

	// Start server
	log.Println("Starting service on port", PORT)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic()
	}
}

// func (app *application) serve() {
// 	srv := &http.Server{
// 		Addr:    fmt.Sprintf(":%s", PORT),
// 		Handler: app.routes(),
// 	}

// 	err := srv.ListenAndServe()
// 	if err != nil {
// 		log.Panic()
// 	}
// }

func connectToMongo() (*mongo.Client, error) {
	// Connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// Connect to DB
	con, err := mongo.Connect(context.TODO(), options.Client())
	if err != nil {
		log.Println("Error connection:", err)
		return nil, err
	}

	log.Println("Connected to Mongo")

	return con, nil
}
