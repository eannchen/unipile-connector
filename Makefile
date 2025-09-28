# Makefile for Unipile Connector

.PHONY: help build run test clean docker-build docker-run migrate dev

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  migrate      - Run database migrations"
	@echo "  dev          - Run in development mode"

# Build the application
build:
	go build -o bin/unipile-connector ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Build Docker image
docker-build:
	docker build -t unipile-connector .

# Run with Docker Compose
docker-up:
	docker compose -f ./docker-compose.yml up -d --build

# Stop docker compose
docker-down:
	docker compose -f ./docker-compose.yml down

# Run database migrations
migrate:
	go run ./cmd/api migrate

# Development mode with hot reload
dev:
	@echo "Starting development server..."
	@echo "Make sure PostgreSQL and Redis are running"
	go run ./cmd/api

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate mocks (if using mockgen)
mocks:
	mockgen -source=internal/domain/repository/user_repository.go -destination=internal/mocks/user_repository_mock.go
	mockgen -source=internal/domain/repository/account_repository.go -destination=internal/mocks/account_repository_mock.go

