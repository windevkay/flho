# Include variables from the .env file
include .env


# ==================================================================================== #
# HELPERS
# ==================================================================================== #


## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'


.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #


## run/api: run the cmd/api application
.PHONY: run/api
run/api: build/api
	@echo 'Running Docker Compose'
	docker compose -f docker-compose.yaml up --build -d
	@echo 'Docker images started!'


## down: stop docker compose
.PHONY: down
down:
	@echo 'Stopping docker compose...'
	docker compose -f docker-compose.yaml down
	@echo 'Done!'


## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${DB_DSN}


## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}


## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${DB_DSN} up


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #


.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...


.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor


# ==================================================================================== #
# BUILD
# ==================================================================================== #


## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api targetting 64 bit linux...'
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api
	@echo 'Binary generation successful!'