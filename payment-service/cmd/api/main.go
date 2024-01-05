package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/AthfanFasee/payment/internal/data"
	"github.com/AthfanFasee/payment/utils"
	_ "github.com/go-sql-driver/mysql"
	amqp "github.com/rabbitmq/amqp091-go"
)

const version = "1.0.0"

type config struct {
	port      int
	env       string
	mysqlDSN  string
	rabbitDSN string
	stripe    struct {
		secret string
		key    string
	}
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	DB       *sql.DB
	Models   data.Models
	RabbitMQ *amqp.Connection
}

// Starts listening to http calls
func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	app.infoLog.Printf("Server started on port %d\n", app.config.port)

	return srv.ListenAndServe()
}

func main() {
	var cfg config

	// Set up configs from env file
	env, err := utils.LoadEnv()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	cfg.port = env.Port
	cfg.mysqlDSN = env.MysqlDSN
	cfg.rabbitDSN = env.RabbitDSN
	cfg.stripe.key = env.StripeKey
	cfg.stripe.secret = env.StripeSecret

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	flag.StringVar(&cfg.env, "env", "development", "Application environment {development|production|maintenance}")

	flag.Parse()

	sqlConn, err := ConnectToMySql(cfg.mysqlDSN)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer sqlConn.Close()

	rabbitConn, err := connectToRabbit(cfg.rabbitDSN)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer rabbitConn.Close()

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		DB:       sqlConn,
		RabbitMQ: rabbitConn,
		Models:   data.NewModels(sqlConn),
	}

	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}
}

// Opens a connection to mySQL
func ConnectToMySql(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return db, nil
}

// Opens a connection to rabbitMQ
func connectToRabbit(dsn string) (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// Wait until rabbitMQ is ready
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
