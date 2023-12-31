package main

import (
	"log"
	"os"
	"strconv"

	util "github.com/AthfanFasee/mail/utils"
)

const PORT = "80"

type config struct {
	mail struct {
		Domain      string
		Host        string
		Port        string
		Username    string
		Password    string
		Encryption  string
		FromName    string
		FromAddress string
	}
	gRPCPort int
}

type application struct {
	config config
	Mailer Mail
}

func main() {
	var cfg config

	// Set up configs from env file
	env, err := util.LoadEnv()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	cfg.mail.Domain = env.Domain
	cfg.mail.Host = env.Host
	cfg.mail.Port = env.Port
	cfg.mail.Username = env.Username
	cfg.mail.Password = env.Password
	cfg.mail.Encryption = env.Encryption
	cfg.mail.FromName = env.FromName
	cfg.mail.FromAddress = env.FromAddress
	cfg.gRPCPort = env.GrpcServerPort

	app := application{
		Mailer: createMail(cfg),
	}

	log.Println("Starting authentication service")

	app.gRPCListen()

}

func createMail(cfg config) Mail {
	port, _ := strconv.Atoi(cfg.mail.Port)
	m := Mail{
		Domain:      cfg.mail.Domain,
		Host:        cfg.mail.Host,
		Port:        port,
		Username:    cfg.mail.Username,
		Password:    cfg.mail.Password,
		Encryption:  cfg.mail.Encryption,
		FromName:    cfg.mail.FromName,
		FromAddress: cfg.mail.FromAddress,
	}

	return m
}
