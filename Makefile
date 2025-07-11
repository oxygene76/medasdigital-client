# Makefile for MedasDigital Client

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary info
BINARY_NAME=medasdigital-client
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe
BINARY_DARWIN=$(BINARY_NAME)_darwin

# Build directory
BUILD_DIR=bin

# Version info
VERSION ?= $(shell git describe --tags --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Linker flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

.PHONY: all build build-linux build-windows build-darwin clean test deps help install run

# Default target
all: clean deps test build

# Build for current OS
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/medasdigital-client

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) ./cmd/medasdigital-client

# Build for Windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) ./cmd/medasdigital-client

# Build for macOS
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN) ./cmd/medasdigital-client

# Build for all platforms
build-all: build-linux build-windows build-darwin

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -cover ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Install binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Initialize client configuration
init-config:
	@echo "Initializing client configuration..."
	./$(BUILD_DIR)/$(BINARY_NAME) init --chain-id medasdigital-2

# Check GPU availability
check-gpu:
	@echo "Checking GPU availability..."
	./$(BUILD_DIR)/$(BINARY_NAME) gpu status

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Lint code
lint:
	@echo "Running linter..."
	golangci-lint run

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	$(GOCMD) doc -all ./... > docs/api-reference.md

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t medasdigital-client:$(VERSION) .

# Docker build with GPU support
docker-build-gpu:
	@echo "Building Docker image with GPU support..."
	docker build -f docker/Dockerfile.gpu -t medasdigital-client:$(VERSION)-gpu .

# Development setup
dev-setup: deps
	@echo "Setting up development environment..."
	@echo "Installing development tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "Creating sample config..."
	mkdir -p configs
	cp configs/config.yaml.example configs/config.yaml || true

# Release build (all platforms)
release: clean test build-all
	@echo "Creating release artifacts..."
	@mkdir -p release
	@cp $(BUILD_DIR)/$(BINARY_UNIX) release/$(BINARY_NAME)-$(VERSION)-linux-amd64
	@cp $(BUILD_DIR)/$(BINARY_WINDOWS) release/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe
	@cp $(BUILD_DIR)/$(BINARY_DARWIN) release/$(BINARY_NAME)-$(VERSION)-darwin-amd64
	@echo "Release artifacts created in release/ directory"

# Quick development test
dev-test: build
	@echo "Running development test..."
	./$(BUILD_DIR)/$(BINARY_NAME) status || echo "Client not initialized"

# Setup test blockchain connection
test-connection: build
	@echo "Testing blockchain connection..."
	./$(BUILD_DIR)/$(BINARY_NAME) status

# Example workflows
example-orbital: build
	@echo "Running example orbital dynamics analysis..."
	./$(BUILD_DIR)/$(BINARY_NAME) analyze orbital-dynamics \
		--input data/samples/tno_elements.csv \
		--output results/orbital_analysis.json

example-gpu-training: build
	@echo "Running example GPU training..."
	./$(BUILD_DIR)/$(BINARY_NAME) train deep-detector \
		--training-data data/samples/training_set.h5 \
		--model-architecture resnet50 \
		--gpu-devices 0 \
		--batch-size 16 \
		--epochs 10

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build for current OS"
	@echo "  build-all      - Build for all platforms"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  deps           - Install dependencies"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  run            - Build and run"
	@echo "  init-config    - Initialize client configuration"
	@echo "  check-gpu      - Check GPU availability"
	@echo "  benchmark      - Run benchmarks"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  security       - Run security scan"
	@echo "  docs           - Generate documentation"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-build-gpu - Build Docker image with GPU support"
	@echo "  dev-setup      - Setup development environment"
	@echo "  release        - Create release build"
	@echo "  dev-test       - Quick development test"
	@echo "  test-connection - Test blockchain connection"
	@echo "  example-orbital - Run orbital dynamics example"
	@echo "  example-gpu-training - Run GPU training example"
	@echo "  help           - Show this help"
