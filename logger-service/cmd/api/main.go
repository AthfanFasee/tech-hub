package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/AthfanFasee/logger-service/data"
	"github.com/AthfanFasee/logger-service/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type config struct {
	gRPCPort int
	mongoURL string
}

type application struct {
	Models data.Models
	config config
}

func main() {
	var cfg config

	// Set up configs from env file
	env, err := utils.LoadEnv()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	cfg.gRPCPort = env.GrpcServerPort
	cfg.mongoURL = env.MongoURL

	log.Println("Starting logger service")

	mongoClient, err := connectToMongo(cfg.mongoURL)
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
		config: cfg,
	}

	// Start gRPC server
	app.gRPCListen()
}

// Opens a connection to mongoDB
func connectToMongo(mongoURL string) (*mongo.Client, error) {
	// Connection options
	clientOptions := options.Client().ApplyURI(mongoURL)
	// Make these come from env variable in prod
	clientOptions.SetAuth(options.Credential{
		Username: "root",
		Password: "secret",
	})

	con, err := mongo.Connect(context.TODO(), options.Client())
	if err != nil {
		log.Println("Error connection:", err)
		return nil, err
	}

	log.Println("Connected to mongoDB")

	return con, nil
}
