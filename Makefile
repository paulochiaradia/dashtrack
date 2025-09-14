# Dashtrack API Makefile

.PHONY: help build test test-unit test-integration test-benchmark clean docker-up docker-down migrate

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build the application
	@echo "Building application..."
	@go build -o bin/dashtrack ./cmd/api

build-docker: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t dashtrack-api .

# Test targets
test: test-unit test-integration test-benchmark ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	@go test -v -race -coverprofile=coverage.out ./internal/...

test-integration: ## Run integration tests  
	@echo "Running integration tests..."
	@go test -v -race ./tests/integration/...

test-benchmark: ## Run benchmark tests
	@echo "Running benchmark tests..."
	@go test -v -bench=. -benchmem ./tests/benchmarks/...

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./internal/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Development targets
run: ## Run the application locally
	@echo "Starting application..."
	@go run ./cmd/api

run-dev: ## Run with live reload (if air is installed)
	@echo "Starting with live reload..."
	@air

# Database targets
migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@go run ./cmd/api migrate up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@go run ./cmd/api migrate down

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	@go run ./cmd/api migrate status

# Docker targets
docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	@docker-compose up -d

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-logs: ## Show Docker logs
	@docker-compose logs -f

docker-rebuild: ## Rebuild and start Docker containers
	@echo "Rebuilding Docker containers..."
	@docker-compose down
	@docker-compose up --build -d

# Database Docker targets
db-reset: ## Reset database (WARNING: This will delete all data!)
	@echo "Resetting database..."
	@docker-compose down
	@docker volume rm dashtrack_postgres_data || true
	@docker-compose up -d db
	@sleep 5
	@docker-compose up api

# Monitoring targets
prometheus: ## Start Prometheus (if docker-compose.monitoring.yml exists)
	@echo "Starting Prometheus..."
	@docker-compose -f docker-compose.monitoring.yml up -d prometheus

grafana: ## Start Grafana (if docker-compose.monitoring.yml exists)
	@echo "Starting Grafana..."
	@docker-compose -f docker-compose.monitoring.yml up -d grafana

monitoring-up: ## Start all monitoring services
	@echo "Starting monitoring stack..."
	@docker-compose -f docker-compose.monitoring.yml up -d

monitoring-down: ## Stop all monitoring services
	@echo "Stopping monitoring stack..."
	@docker-compose -f docker-compose.monitoring.yml down

# Code quality targets
lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning up..."
	@rm -f bin/dashtrack
	@rm -f coverage.out
	@rm -f coverage.html
	@go clean -cache

# Security targets
security-scan: ## Run security scan with gosec
	@echo "Running security scan..."
	@gosec ./...

# Load testing targets
load-test: ## Run basic load test (requires hey tool)
	@echo "Running load test on health endpoint..."
	@hey -n 1000 -c 10 http://localhost:8080/health

# API testing targets
api-test-health: ## Test health endpoint
	@echo "Testing health endpoint..."
	@curl -s http://localhost:8080/health | jq

api-test-roles: ## Test roles endpoint
	@echo "Testing roles endpoint..."
	@curl -s http://localhost:8080/roles | jq

api-test-users: ## Test users endpoint
	@echo "Testing users endpoint..."
	@curl -s http://localhost:8080/users | jq

# Setup targets
setup: ## Install development dependencies
	@echo "Installing development dependencies..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Info targets
info: ## Show project information
	@echo "Project: Dashtrack API"
	@echo "Go version: $(shell go version)"
	@echo "Module: $(shell go list -m)"
	@echo "Dependencies: $(shell go list -m all | wc -l) modules"
