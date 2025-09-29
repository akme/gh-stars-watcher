#!/usr/bin/env bash
set -euo pipefail

# Go tooling configuration for GitHub Stars Watcher

# Format all Go code
format() {
    echo "Running gofmt..."
    gofmt -w .
    echo "Running goimports..."
    go run golang.org/x/tools/cmd/goimports@latest -w .
}

# Lint Go code
lint() {
    echo "Running go vet..."
    go vet ./...
    echo "Running golint..."
    go run golang.org/x/lint/golint@latest ./...
}

# Run all tests
test() {
    echo "Running tests..."
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
}

# Build the application
build() {
    echo "Building star-watcher..."
    go build -o bin/star-watcher ./cmd/star-watcher
}

# Clean build artifacts
clean() {
    echo "Cleaning build artifacts..."
    rm -rf bin/ coverage.out coverage.html
}

# Install development dependencies
dev-deps() {
    echo "Installing development dependencies..."
    go install golang.org/x/tools/cmd/goimports@latest
    go install golang.org/x/lint/golint@latest
}

# Show help
help() {
    echo "Available commands:"
    echo "  format    - Format Go code with gofmt and goimports"
    echo "  lint      - Lint Go code with go vet and golint"
    echo "  test      - Run all tests with coverage"
    echo "  build     - Build the application"
    echo "  clean     - Clean build artifacts"
    echo "  dev-deps  - Install development dependencies"
    echo "  help      - Show this help message"
}

# Main command dispatcher
case "${1:-help}" in
    format) format ;;
    lint) lint ;;
    test) test ;;
    build) build ;;
    clean) clean ;;
    dev-deps) dev-deps ;;
    help) help ;;
    *) echo "Unknown command: $1"; help; exit 1 ;;
esac