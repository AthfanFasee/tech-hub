package main

import (
	"fmt"
	"log"
	"net/http"
)

const PORT = "80"

type application struct {
}

func main() {
	app := application{}

	log.Printf("Server started on port %s\n", PORT)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}
}
