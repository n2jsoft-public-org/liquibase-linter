#!/bin/bash
# Makefile for Liquibase Linter

# Variables
BINARY_NAME=liquibase-linter
BUILD_DIR=build
VERSION=dev

# Build the application
build:
	go build -ldflags "-X main.version=$(VERSION)" -o "$(BUILD_DIR)/$(BINARY_NAME)" ./cmd/liquibase-linter

# Run tests
test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run linter
lint:
	golangci-lint run ./...

# Clean up build artifacts
clean:
	go clean

# Coverage summary
coverage:
	go tool cover -func=coverage.txt | grep total

# View detailed coverage in browser
view-coverage:
	go tool cover -html=coverage.txt

.PHONY: build test lint clean coverage view-coverage