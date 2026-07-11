MODULE      := github.com/Djarottosca/finalproject-ftgo-1
MIGRATIONS  := migrations
# DB_URL, e.g. postgres://user:pass@localhost:5432/dbname?sslmode=disable
DB_URL      ?=

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

## --- Run ---------------------------------------------------------------

.PHONY: run-core run-worker run-payment run-notification
run-core: ## Run core-service HTTP server
	go run ./core-service/cmd/server

run-worker: ## Run core-service asynq worker
	go run ./core-service/cmd/worker

run-payment: ## Run payment-service
	go run ./payment-service/cmd/server

run-notification: ## Run notification-service
	go run ./notification-service/cmd/server

## --- Build ---------------------------------------------------------------

.PHONY: build
build: ## Build all service binaries into ./bin
	go build -o bin/core-server ./core-service/cmd/server
	go build -o bin/core-worker ./core-service/cmd/worker
	go build -o bin/payment-server ./payment-service/cmd/server
	go build -o bin/notification-server ./notification-service/cmd/server

## --- Quality ---------------------------------------------------------------

.PHONY: fmt lint vet tidy test
fmt: ## Format code and sort imports (gofmt + goimports-reviser)
	gofmt -w -s .
	@which goimports-reviser > /dev/null || go install github.com/incu6us/goimports-reviser/v3@latest
	goimports-reviser -project-name $(MODULE) -rm-unused -set-alias ./...

lint: fmt vet ## Run fmt + vet

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy go.mod/go.sum
	go mod tidy

test: ## Run tests
	go test ./... -v

## --- Database migrations ---------------------------------------------------

.PHONY: migrate-up migrate-down migrate-force migrate-version migrate-create
migrate-up: ## Apply all up migrations (requires DB_URL)
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" up

migrate-down: ## Roll back one migration (requires DB_URL)
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" down 1

migrate-force: ## Force migration version, usage: make migrate-force VERSION=1
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" force $(VERSION)

migrate-version: ## Print current migration version
	migrate -path $(MIGRATIONS) -database "$(DB_URL)" version

migrate-create: ## Create a new migration pair, usage: make migrate-create NAME=add_something
	migrate create -ext sql -dir $(MIGRATIONS) -seq $(NAME)

## --- Protobuf ---------------------------------------------------------------

.PHONY: proto
proto: ## Generate gRPC code from proto/ via buf
	buf generate

## --- Misc ---------------------------------------------------------------

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf bin
