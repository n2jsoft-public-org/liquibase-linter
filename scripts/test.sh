#!/bin/bash
# Test script for Liquibase Linter

set -e

echo "Running tests with coverage..."
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

echo ""
echo "Coverage summary:"
go tool cover -func=coverage.txt | grep total

echo ""
echo "To view detailed coverage in browser, run:"
echo "  go tool cover -html=coverage.txt"
