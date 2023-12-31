package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/AthfanFasee/payment/internal/data"
	"github.com/AthfanFasee/payment/utils"
	_ "github.com/go-sql-driver/mysql"
)

const version = "1.0.0"

type config struct {
	port     int
	env      string
	mysqlDSN string
	stripe   struct {
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

	app.infoLog.Printf("Starting Back end server in %s mode on port %d\n", app.config.env, app.config.port)

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
	cfg.stripe.key = env.StripeKey
	cfg.stripe.secret = env.StripeSecret

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	flag.StringVar(&cfg.env, "env", "development", "Application environment {development|production|maintenance}")

	flag.Parse()

	db, err := OpenDB(cfg.mysqlDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		DB:       db,
		Models:   data.NewModels(db),
	}

	err = app.serve()
	if err != nil {
		log.Fatal(err)
	}
}

// Connect to MariaDB
func OpenDB(dsn string) (*sql.DB, error) {
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
