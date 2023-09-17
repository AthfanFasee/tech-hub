package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

const PORT = "80"

type application struct {
	Mailer Mail
}

func main() {
	app := application{
		Mailer: createMail(),
	}

	log.Println("Starting main service on port", PORT)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}

}

func createMail() Mail {
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	m := Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Host:        os.Getenv("MAIL_HOST"),
		Port:        port,
		Username:    os.Getenv("MAIL_USERNAME"),
		Password:    os.Getenv("MAIL_PASSWORD"),
		Encryption:  os.Getenv("MAIL_ENCRYPTION"),
		FromName:    os.Getenv("MAIL_NAME"),
		FromAddress: os.Getenv("MAIL_ADDRESS"),
	}

	return m
}
