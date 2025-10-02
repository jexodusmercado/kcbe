# Makefile for kabancount-be

.PHONY: help build run test clean dev docker-up docker-down docker-reset install-air setup-dev tidy lint

# Default target
.DEFAULT_GOAL := help

# Go parameters
BINARY_NAME=kabancount-be
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./main.go
PORT=8080

## help: Show this help message
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Built $(BINARY_PATH)"

## run: Run the application directly
run:
	@echo "Running $(BINARY_NAME) on port $(PORT)..."
	@go run $(MAIN_PATH) -port=$(PORT)

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Clean build artifacts and temporary files
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -rf tmp/
	@rm -f coverage.out coverage.html
	@rm -f build-errors.log

## dev: Start development server with hot reload using Air
dev: install-air
	@echo "Starting development server with Air..."
	@air

## install-air: Install Air for hot reloading if not already installed
install-air:
	@if ! command -v air > /dev/null; then \
		echo "Installing Air..."; \
		go install github.com/air-verse/air@latest; \
	fi

## docker-up: Start PostgreSQL databases using docker-compose
docker-up:
	@echo "Starting PostgreSQL databases..."
	@docker-compose up -d

## docker-down: Stop PostgreSQL databases
docker-down:
	@echo "Stopping PostgreSQL databases..."
	@docker-compose down

## docker-reset: Reset PostgreSQL databases (stop, remove volumes, and start fresh)
docker-reset:
	@echo "Resetting PostgreSQL databases..."
	@docker-compose down -v
	@docker-compose up -d

## docker-logs: View database logs
docker-logs:
	@docker-compose logs -f db

## db-connect: Connect to main database via psql
db-connect:
	@echo "Connecting to main database..."
	@docker exec -it kabancount_db psql -U postgres -d kabancount

## db-connect-test: Connect to test database via psql
db-connect-test:
	@echo "Connecting to test database..."
	@docker exec -it kabancount_test_db psql -U postgres -d kabancount_test

## tidy: Clean up Go modules
tidy:
	@echo "Tidying Go modules..."
	@go mod tidy

## lint: Run linter (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## setup-dev: Set up development environment
setup-dev: install-air docker-up tidy
	@echo "Development environment setup complete!"
	@echo "Run 'make dev' to start development server"

## all: Build and test (default target for CI/CD)
all: clean tidy build test

# Development shortcuts
.PHONY: start stop restart
## start: Alias for 'make dev'
start: dev

## stop: Stop any running Air processes
stop:
	@pkill -f "air" || true
	@echo "Stopped development server"

## restart: Restart development server
restart: stop dev