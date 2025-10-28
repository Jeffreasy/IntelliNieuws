# Makefile for Nieuws Scraper

.PHONY: help build run test clean fmt lint tidy dev-setup

# Variables
APP_NAME=nieuws-scraper
API_BINARY=bin/api
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")

# Default target
help:
	@echo "Available targets:"
	@echo "  make build       - Build the API binary"
	@echo "  make run         - Run the API locally"
	@echo "  make test        - Run tests"
	@echo "  make test-cover  - Run tests with coverage"
	@echo "  make fmt         - Format Go code"
	@echo "  make lint        - Run linters"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make tidy        - Tidy go modules"
	@echo "  make dev-setup   - Setup development environment"

# Build the API binary
build:
	@echo "Building API..."
	@mkdir -p bin
	@go build -o $(API_BINARY) ./cmd/api
	@echo "Build complete: $(API_BINARY)"

# Run the API locally
run:
	@echo "Running API..."
	@go run ./cmd/api/main.go

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

# Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run
	@echo "Linting complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Tidy go modules
tidy:
	@echo "Tidying modules..."
	@go mod tidy
	@echo "Modules tidied"

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@if not exist .env copy .env.example .env
	@echo "Created .env file"
	@go mod download
	@echo "Downloaded dependencies"
	@echo ""
	@echo "Setup complete!"
	@echo ""
	@echo "IMPORTANT: Install PostgreSQL and Redis locally:"
	@echo "  1. PostgreSQL: https://www.postgresql.org/download/windows/"
	@echo "  2. Redis: https://github.com/microsoftarchive/redis/releases"
	@echo ""
	@echo "Then run: make run"