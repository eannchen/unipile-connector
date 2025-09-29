# Makefile for Unipile Connector

.PHONY: help build run test clean docker-up docker-down fmt

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-up    - Run with Docker Compose for development"
	@echo "  docker-down  - Stop with Docker Compose for development"
	@echo "  fmt          - Format code"

# Build the application
build:
	go build -o bin/unipile-connector ./cmd/api

# Run the application
run:
	go run ./cmd/api

# Run tests
test:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Run with Docker Compose for development
docker-up:
	docker compose -f ./docker-compose.yml up -d --build

# Stop docker compose for development
docker-down:
	docker compose -f ./docker-compose.yml down

# Format code
fmt:
	go fmt ./...