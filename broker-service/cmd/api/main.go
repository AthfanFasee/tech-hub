package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Configuration settings
type config struct {
	port    int
	env     string
	metrics bool
	dsn     string
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	Rabbit *amqp.Connection
}

func main() {
	var cfg config

	cfg.dsn = "amqp://guest:guest@rabbitmq"
	cfg.port = 80

	// Server Related
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.BoolVar(&cfg.metrics, "metrics", false, "Enable metrics")
	// Rate limiting Related
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	// Cors related
	flag.Func("cors-trusted-origins", "Trusted CORS origins(separated by space)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	// Connect to rabbitmq.
	rabbitConn, err := connectToRabbit(cfg.dsn)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer rabbitConn.Close()

	app := application{
		config: cfg,
		Rabbit: rabbitConn,
	}

	log.Printf("Server started on port %d\n", cfg.port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
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
