# Makefile for Liquibase Linter

# Variables
BINARY_NAME=liquibase-linter

# Build the application
build:
	go build -o $(BINARY_NAME) ./cmd/liquibase-linter

# Run tests
test:
	go test ./...

# Run tests with coverage
coverage:
	go test -cover ./...

# Run linters
lint:
	golangci-lint run

# Clean up build artifacts
clean:
	go clean

depends: build test lint

.PHONY: build test coverage lint clean depends