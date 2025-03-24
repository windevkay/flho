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
run/api:
	@echo 'Running Docker Compose'
	docker compose -f docker-compose.yml up --build -d
	@echo 'Docker images started!'


## down: stop docker compose
.PHONY: down
down:
	@echo 'Stopping docker compose...'
	docker compose -f docker-compose.yml down
	@echo 'Done!'


.PHONY: start-test-resources
start-test-resources:
	@echo 'Starting test resources...'
	docker compose -f docker-compose-testing.yml up --build -d
	@echo 'Docker images started!'


## tests: run application tests
.PHONY: tests
tests: start-test-resources
	@echo 'Running tests...'
	go test -race -vet=off ./...
	@echo 'Tests completed - tearing down test resources'
	docker compose -f docker-compose-testing.yml down
	@echo 'Docker images stopped!'
	@echo 'Pruning volumes...'
	docker system prune -f
	@echo 'Done!'


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


.PHONY: swagger
swagger:
	~/go/bin/swag init -g cmd/api/main.go -q --parseDependency --parseInternal -o ./docs


.PHONY: swag
swag:
	go install github.com/swaggo/swag/cmd/swag@latest

