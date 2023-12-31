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

	util "github.com/AthfanFasee/broker/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Configuration settings
type config struct {
	port      int
	env       string
	metrics   bool
	rabbitDSN string
	limiter   struct {
		rps     float64
		burst   int
		enabled bool
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config   config
	RabbitMQ *amqp.Connection
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	log.Printf("Server started on port %d\n", app.config.port)

	return srv.ListenAndServe()
}

func main() {
	var cfg config

	// Set up configs from env file
	env, err := util.LoadEnv()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	cfg.port = env.Port
	cfg.rabbitDSN = env.RabbitDSN

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
	rabbitConn, err := connectToRabbit(cfg.rabbitDSN)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer rabbitConn.Close()

	app := application{
		config:   cfg,
		RabbitMQ: rabbitConn,
	}

	err = app.serve()
	if err != nil {
		log.Fatal(err)
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
