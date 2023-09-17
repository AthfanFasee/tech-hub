SHELL=cmd.exe
FRONT_END_BINARY=frontApp.exe
BROKER_BINARY=brokerApp
AUTH_BINARY=authenticationApp
LOGGER_BINARY=loggerApp
MAIL_BINARY=mailerApp
LISTENER_BINARY=listenerApp


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## up: starts all containers in the background without forcing build
.PHONY: up
up:
	@echo Starting Docker images...
	docker-compose up -d
	@echo Docker images started!

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
.PHONY: up_build
up_build: build_broker build_auth build_logger build_mail build_listener
	@echo Stopping docker images (if running...)
	docker-compose down
	@echo Building (when required) and starting docker images...
	docker-compose up --build -d
	@echo Docker images built and started!

## down: stop docker compose
.PHONY: down
down:
	@echo Stopping docker compose...
	docker-compose down
	@echo Done!

## build_broker: builds the broker binary as a linux executable
.PHONY: build_broker
build_broker:
	@echo Building broker binary...
	chdir ..\broker-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${BROKER_BINARY} ./cmd/api
	@echo Done!

## build_logger: builds the broker binary as a linux executable
.PHONY: build_logger
build_logger:
	@echo Building logger binary...
	chdir ..\logger-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${LOGGER_BINARY} ./cmd/api
	@echo Done!

## build_auth: builds the auth binary as a linux executable
.PHONY: build_auth
build_auth:
	@echo Building auth binary...
	chdir ..\authentication-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${AUTH_BINARY} ./cmd/api
	@echo Done!

## build_mail: builds the auth binary as a linux executable
.PHONY: build_mail
build_mail:
	@echo Building mail binary...
	chdir ..\mail-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${MAIL_BINARY} ./cmd/api
	@echo Done!

## build_listener: builds the auth binary as a linux executable
.PHONY: build_listener
build_listener:
	@echo Building listener binary...
	chdir ..\listener-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0 && go build -o ${LISTENER_BINARY} .
	@echo Done!
