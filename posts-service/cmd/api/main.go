package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/AthfanFasee/posts/internal/data"
	"github.com/AthfanFasee/posts/util"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type config struct {
	gRPCPort   int
	postgreDSN string
	rabbitDSN  string
}

type application struct {
	config config
	DB     *sql.DB
	Models data.Models
	Rabbit *amqp.Connection
}

func main() {
	var cfg config

	// Set up configs from env file
	env, err := util.LoadEnv()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	cfg.gRPCPort = env.GrpcServerPort
	cfg.postgreDSN = env.PostgreDSN
	cfg.rabbitDSN = env.RabbitDSN

	log.Println("Starting authentication service")

	// Connect to PostgreSQL
	db, err := connectToPostgreSQL(cfg.postgreDSN)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	// Connect to RabbitMQ
	rabbitConn, err := connectToRabbit(cfg.rabbitDSN)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	if db == nil {
		log.Panic()
	}

	defer db.Close()

	app := &application{
		config: cfg,
		DB:     db,
		Models: data.NewModels(db),
		Rabbit: rabbitConn,
	}

	// Start gRPC server
	app.gRPCListen()
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToPostgreSQL(dsn string) (*sql.DB, error) {
	var counts int64

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready ...")
			counts++
		} else {
			log.Println("Connected to Postgres!")
			return connection, nil
		}

		if counts > 10 {
			return nil, err
		}

		log.Println("Backing off for two seconds....")
		time.Sleep(2 * time.Second)
		continue
	}
}

func connectToRabbit(dsn string) (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// Wait until rabbitmq is ready
	for {
		c, err := amqp.Dial(dsn)
		if err != nil {
			fmt.Println("RabbitMQ is not yet ready...")
			counts++
		} else {
			log.Println("Connected to RabbitMQ")
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("Backing off...")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
