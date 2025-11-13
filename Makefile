.PHONY: help build run test test-coverage clean docker-build docker-run docker-down lint migrate-local migrate-remote create-migration

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building application..."
	@CGO_ENABLED=1 go build -o bin/api cmd/api/main.go

run: ## Run the application locally
	@echo "Running application..."
	@go run cmd/api/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -f *.db *.db-shm *.db-wal

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t cloudflaredb-api:latest .

docker-run: ## Run application with Docker Compose
	@echo "Starting application with Docker Compose..."
	@docker-compose up -d

docker-down: ## Stop Docker Compose
	@echo "Stopping Docker Compose..."
	@docker-compose down

docker-logs: ## View Docker logs
	@docker-compose logs -f

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

migrate-local: ## Run migrations on local D1 database
	@echo "Running migrations locally..."
	@./scripts/run-migrations.sh local

migrate-remote: ## Run migrations on remote D1 database (production)
	@echo "Running migrations on production..."
	@./scripts/run-migrations.sh remote

create-migration: ## Create a new migration file
	@./scripts/create-migration.sh

.DEFAULT_GOAL := help
