.PHONY: help build run test clean install deps

help:
	@echo "GO Money - CLI for managing expenses"
	@echo ""
	@echo "Available commands:"
	@echo "  make build    - Build the binary"
	@echo "  make run      - Run the application"
	@echo "  make test     - Run tests"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make install  - Install dependencies"
	@echo "  make deps     - Download dependencies"
	@echo "  make fmt      - Format code"
	@echo "  make lint     - Run linter"

build:
	@echo "Building GO Money..."
	@go build -o bin/gm ./cmd/main.go

run:
	@echo "Running GO Money..."
	@go run ./cmd/main.go

test:
	@echo "Running tests..."
	@go test -v ./...

clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@go clean

install:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

deps:
	@echo "Downloading dependencies..."
	@go mod download

fmt:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

.DEFAULT_GOAL := help
