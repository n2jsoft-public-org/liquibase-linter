#!/bin/bash
# Build script for Liquibase Linter

set -e

VERSION="${VERSION:-dev}"
BUILD_DIR="build"
BINARY_NAME="liquibase-linter"

echo "Building Liquibase Linter version ${VERSION}..."

# Create build directory
mkdir -p "${BUILD_DIR}"

# Build for current platform
echo "Building for current platform..."
go build -ldflags "-X main.version=${VERSION}" -o "${BUILD_DIR}/${BINARY_NAME}" ./cmd/liquibase-linter

echo "Build complete: ${BUILD_DIR}/${BINARY_NAME}"
echo "Run './${BUILD_DIR}/${BINARY_NAME} version' to verify"
