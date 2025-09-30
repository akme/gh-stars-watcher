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

# Docker build targets
docker-build:
	UPX_VERSION=${UPX_VERSION} ./scripts/docker-build.sh local

docker-build-multiarch:
	UPX_VERSION=${UPX_VERSION} ./scripts/docker-build.sh test

docker-push:
	UPX_VERSION=${UPX_VERSION} ./scripts/docker-build.sh push

docker-push-version:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make docker-push-version VERSION=v1.0.0"; exit 1; fi
	UPX_VERSION=${UPX_VERSION} ./scripts/docker-build.sh push $(VERSION)

# Docker development
docker-dev:
	docker-compose up star-watcher-dev

docker-run:
	docker-compose up star-watcher

# Docker cleanup
docker-clean:
	docker-compose down -v
	docker buildx prune -f

# Show help
help:
	@echo "Available commands:"
	@echo "  build              Build the application"
	@echo "  test               Run all tests"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  clean              Clean build artifacts"
	@echo "  deps               Install dependencies"
	@echo "  lint               Run linter"
	@echo "  format             Format code"
	@echo "  install            Install binary to /usr/local/bin"
	@echo "  docker-build       Build Docker image locally"
	@echo "  docker-build-multiarch  Test multi-arch build"
	@echo "  docker-push        Build and push to registry"
	@echo "  docker-push-version    Push with version tag (requires VERSION=)"
	@echo "  docker-dev         Run development container"
	@echo "  docker-run         Run production container"
	@echo "  docker-clean       Clean Docker resources"
	@echo "  help               Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  UPX_VERSION        UPX version for Docker builds (default: 5.0.2)"
	@echo "  VERSION            Version tag for docker-push-version"

# Default target
all: format lint test build