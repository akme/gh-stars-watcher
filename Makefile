# GitHub Stars Monitor Makefile

.PHONY: build test clean install lint format help

# Build the application
build:
	go build -o bin/star-watcher ./cmd/star-watcher

# Run all tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f bin/star-watcher
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Format code
format:
	go fmt ./...
	go mod tidy

# Install the binary
install: build
	cp bin/star-watcher /usr/local/bin/

# Show help
help:
	@echo "Available commands:"
	@echo "  build          Build the application"
	@echo "  test           Run all tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  clean          Clean build artifacts"
	@echo "  deps           Install dependencies"
	@echo "  lint           Run linter"
	@echo "  format         Format code"
	@echo "  install        Install binary to /usr/local/bin"
	@echo "  help           Show this help message"

# Default target
all: format lint test build