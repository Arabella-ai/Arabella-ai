# Arabella Backend Makefile

# Variables
APP_NAME := arabella
BINARY := bin/api
MAIN_PATH := ./cmd/api
GO := go
DOCKER_COMPOSE := docker compose
MIGRATE := migrate
SWAG := swag

# Build info
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: all build run clean test lint fmt swagger migrate-up migrate-down docker-up docker-down help

# Default target
all: build

## Build Commands
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@mkdir -p bin
	$(GO) build $(LDFLAGS) -o $(BINARY) $(MAIN_PATH)
	@echo "Build complete: $(BINARY)"

build-linux: ## Build for Linux
	@echo "Building $(APP_NAME) for Linux..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY)-linux $(MAIN_PATH)

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	$(GO) run $(MAIN_PATH)

dev: ## Run with hot-reload (requires air)
	@echo "Running $(APP_NAME) in development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		$(GO) run $(MAIN_PATH); \
	fi

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf tmp/
	@echo "Clean complete"

## Test Commands
test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-short: ## Run short tests
	$(GO) test -v -short ./...

## Code Quality
lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	$(GO) fmt ./...
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	fi

vet: ## Run go vet
	$(GO) vet ./...

## Swagger Documentation
swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@if command -v swag > /dev/null; then \
		$(SWAG) init -g $(MAIN_PATH)/main.go -o docs/; \
	else \
		echo "swag not installed. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

## Database Migrations
migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	$(MIGRATE) -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-down: ## Rollback last migration
	@echo "Rolling back migration..."
	$(MIGRATE) -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down 1

migrate-create: ## Create new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	$(MIGRATE) create -ext sql -dir migrations -seq $(NAME)

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	@echo "Forcing migration version: $(VERSION)"
	$(MIGRATE) -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" force $(VERSION)

## Docker Commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-up: ## Start all containers
	@echo "Starting containers..."
	$(DOCKER_COMPOSE) up -d

docker-up-dev: ## Start containers with dev tools
	@echo "Starting containers with dev tools..."
	$(DOCKER_COMPOSE) --profile dev-tools up -d

docker-down: ## Stop all containers
	@echo "Stopping containers..."
	$(DOCKER_COMPOSE) down

docker-logs: ## View container logs
	$(DOCKER_COMPOSE) logs -f

docker-clean: ## Remove containers and volumes
	@echo "Cleaning up containers and volumes..."
	$(DOCKER_COMPOSE) down -v --remove-orphans

## Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

deps-tidy: ## Tidy dependencies
	$(GO) mod tidy

## Install Tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Tools installed successfully"

## Help
help: ## Show this help
	@echo "Arabella Backend - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Environment variables with defaults
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= arabella
DB_PASSWORD ?= arabella_secret
DB_NAME ?= arabella

