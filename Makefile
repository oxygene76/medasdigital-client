# Makefile for MedasDigital Client
# Enhanced version with payment services and improved features

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

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
MAGENTA=\033[0;35m
CYAN=\033[0;36m
NC=\033[0m # No Color

.PHONY: all build build-linux build-windows build-darwin clean test deps help install run

# Default target
all: clean deps test build

# Build for current OS
build:
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/medasdigital-client
	@echo "$(GREEN)✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Build for Linux
build-linux:
	@echo "$(BLUE)Building for Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_UNIX) ./cmd/medasdigital-client
	@echo "$(GREEN)✅ Linux build completed$(NC)"

# Build for Windows
build-windows:
	@echo "$(BLUE)Building for Windows...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_WINDOWS) ./cmd/medasdigital-client
	@echo "$(GREEN)✅ Windows build completed$(NC)"

# Build for macOS
build-darwin:
	@echo "$(BLUE)Building for macOS...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_DARWIN) ./cmd/medasdigital-client
	@echo "$(GREEN)✅ macOS build completed$(NC)"

# Build for all platforms
build-all: build-linux build-windows build-darwin
	@echo "$(GREEN)✅ All platform builds completed$(NC)"

# Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf release/
	@echo "$(GREEN)✅ Cleanup completed$(NC)"

# Run tests
test:
	@echo "$(BLUE)Running tests...$(NC)"
	$(GOTEST) -v ./...
	@echo "$(GREEN)✅ Tests completed$(NC)"

# Run tests with coverage
test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	$(GOTEST) -v -cover -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(NC)"

# Install dependencies
deps:
	@echo "$(BLUE)Installing dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

# Install binary to GOPATH/bin
install: build
	@echo "$(BLUE)Installing $(BINARY_NAME) to GOPATH/bin...$(NC)"
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "$(GREEN)✅ Installation completed$(NC)"

# Run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# ============================================================================
# MedasDigital Client Specific Commands
# ============================================================================

# Initialize client configuration
init-config: build
	@echo "$(BLUE)Initializing client configuration...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) init
	@echo "$(GREEN)✅ Configuration initialized$(NC)"

# Create demo keys
demo-keys: build
	@echo "$(BLUE)Creating demo keys...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) keys add demo-user || echo "$(YELLOW)Key already exists$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) keys list
	@echo "$(GREEN)✅ Demo keys ready$(NC)"

# Register demo client
demo-register: build demo-keys
	@echo "$(BLUE)Registering demo client...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) register --from demo-user || echo "$(YELLOW)Already registered$(NC)"
	@echo "$(GREEN)✅ Demo client registered$(NC)"

# Check client status
status: build
	@echo "$(BLUE)Checking client status...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) status

# Show current client identity
whoami: build
	@echo "$(BLUE)Checking client identity...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) whoami

# Check account balance
balance: build
	@echo "$(BLUE)Checking account balance...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) balance --from demo-user || echo "$(YELLOW)No balance or key not found$(NC)"

# ============================================================================
# Computing Services
# ============================================================================

# Start free PI computation service
start-free-service: build
	@echo "$(BLUE)Starting free PI computation service...$(NC)"
	@echo "$(YELLOW)Service will run on http://localhost:8080$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) serve --port 8080 --max-jobs 2

# Start payment-enabled service (requires valid addresses)
start-payment-service: build
	@echo "$(BLUE)Starting payment-enabled service...$(NC)"
	@echo "$(RED)⚠️  Requires valid MEDAS addresses!$(NC)"
	@echo "$(YELLOW)Edit this target with your actual addresses$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) payment-service \
		--service-address medas1your-service-address-here \
		--community-address medas1your-community-address-here \
		--port 8080 || echo "$(RED)❌ Failed - update addresses in Makefile$(NC)"

# Test PI calculation directly
test-pi: build
	@echo "$(BLUE)Testing PI calculation (100 digits)...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) pi calculate 100 --method chudnovsky
	@echo "$(GREEN)✅ PI calculation completed$(NC)"

# Run PI benchmark
benchmark-pi: build
	@echo "$(BLUE)Running PI benchmark...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) pi benchmark

# Test free service via curl
test-free-service:
	@echo "$(BLUE)Testing free service API...$(NC)"
	@echo "$(YELLOW)Make sure free service is running first: make start-free-service$(NC)"
	curl -X POST http://localhost:8080/api/v1/calculate \
		-H 'Content-Type: application/json' \
		-d '{"digits": 50, "method": "chudnovsky"}' | jq . || echo "$(RED)Service not running or jq not installed$(NC)"

# ============================================================================
# Original Commands (Enhanced)
# ============================================================================

# Check GPU availability
check-gpu: build
	@echo "$(BLUE)Checking GPU availability...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) gpu status

# Run benchmarks
benchmark:
	@echo "$(BLUE)Running Go benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

# Format code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOCMD) fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

# Lint code
lint:
	@echo "$(BLUE)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)Installing golangci-lint...$(NC)"; \
		$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi
	@echo "$(GREEN)✅ Linting completed$(NC)"

# Security scan
security:
	@echo "$(BLUE)Running security scan...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)Installing gosec...$(NC)"; \
		$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

# Generate documentation
docs:
	@echo "$(BLUE)Generating documentation...$(NC)"
	@mkdir -p docs
	$(GOCMD) doc -all ./... > docs/api-reference.md
	@echo "$(GREEN)✅ Documentation generated: docs/api-reference.md$(NC)"

# Docker build
docker-build:
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t medasdigital-client:$(VERSION) .
	docker tag medasdigital-client:$(VERSION) medasdigital-client:latest
	@echo "$(GREEN)✅ Docker image built: medasdigital-client:$(VERSION)$(NC)"

# Docker build with GPU support
docker-build-gpu:
	@echo "$(BLUE)Building Docker image with GPU support...$(NC)"
	docker build -f docker/Dockerfile.gpu -t medasdigital-client:$(VERSION)-gpu .
	@echo "$(GREEN)✅ GPU Docker image built$(NC)"

# Development setup
dev-setup: deps
	@echo "$(BLUE)Setting up development environment...$(NC)"
	@echo "Installing development tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "Creating sample config..."
	mkdir -p configs
	@echo "$(GREEN)✅ Development environment ready$(NC)"

# Release build (all platforms)
release: clean test build-all
	@echo "$(BLUE)Creating release artifacts...$(NC)"
	@mkdir -p release
	@cp $(BUILD_DIR)/$(BINARY_UNIX) release/$(BINARY_NAME)-$(VERSION)-linux-amd64
	@cp $(BUILD_DIR)/$(BINARY_WINDOWS) release/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe
	@cp $(BUILD_DIR)/$(BINARY_DARWIN) release/$(BINARY_NAME)-$(VERSION)-darwin-amd64
	@echo "$(GREEN)✅ Release artifacts created in release/ directory$(NC)"

# Quick development test
dev-test: build
	@echo "$(BLUE)Running development test...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) status || echo "$(YELLOW)Client not initialized - run 'make init-config'$(NC)"

# Setup test blockchain connection
test-connection: build
	@echo "$(BLUE)Testing blockchain connection...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) status

# ============================================================================
# Complete Demo Workflows
# ============================================================================

# Complete demo setup
demo-setup: init-config demo-keys demo-register
	@echo "$(GREEN)🎉 Demo setup completed!$(NC)"
	@echo "$(CYAN)Next steps:$(NC)"
	@echo "  make status          - Check client status"
	@echo "  make test-pi         - Test PI calculation"
	@echo "  make start-free-service - Start computing service"

# Full development cycle
dev: clean deps test build demo-setup
	@echo "$(GREEN)🚀 Full development cycle completed!$(NC)"

# Quick test of all major features
test-all: build test-pi benchmark-pi
	@echo "$(GREEN)✅ All feature tests completed$(NC)"

# ============================================================================
# Example Workflows (Enhanced)
# ============================================================================

example-orbital: build
	@echo "$(BLUE)Running example orbital dynamics analysis...$(NC)"
	@mkdir -p data/samples results
	@echo "$(YELLOW)Note: Requires sample data in data/samples/$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) analyze orbital-dynamics \
		--input data/samples/tno_elements.csv \
		--output results/orbital_analysis.json || echo "$(YELLOW)Sample data not found$(NC)"

example-gpu-training: build
	@echo "$(BLUE)Running example GPU training...$(NC)"
	@mkdir -p data/samples
	@echo "$(YELLOW)Note: Requires training data and GPU$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) ai train \
		data/samples/training_set.h5 \
		resnet50 \
		--gpu-devices 0 \
		--batch-size 16 \
		--epochs 10 || echo "$(YELLOW)GPU or training data not available$(NC)"

# ============================================================================
# Information and Help
# ============================================================================

# Show build info
info:
	@echo "$(CYAN)MedasDigital Client Build Info$(NC)"
	@echo "$(CYAN)==============================$(NC)"
	@echo "Version:     $(GREEN)$(VERSION)$(NC)"
	@echo "Commit:      $(GREEN)$(COMMIT)$(NC)"
	@echo "Build Date:  $(GREEN)$(DATE)$(NC)"
	@echo "Go Version:  $(GREEN)$(shell $(GOCMD) version)$(NC)"

# Enhanced help with categories
help:
	@echo "$(CYAN)MedasDigital Client Makefile$(NC)"
	@echo "$(CYAN)=============================$(NC)"
	@echo ""
	@echo "$(YELLOW)🔨 Build Commands:$(NC)"
	@echo "  build          - Build for current OS"
	@echo "  build-all      - Build for all platforms"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo ""
	@echo "$(YELLOW)🧪 Testing & Quality:$(NC)"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  benchmark      - Run Go benchmarks"
	@echo "  lint           - Run linter"
	@echo "  security       - Run security scan"
	@echo "  fmt            - Format code"
	@echo ""
	@echo "$(YELLOW)🚀 Demo & Setup:$(NC)"
	@echo "  demo-setup     - Complete demo setup"
	@echo "  dev            - Full development cycle"
	@echo "  init-config    - Initialize client configuration"
	@echo "  demo-keys      - Create demo keys"
	@echo "  demo-register  - Register demo client"
	@echo ""
	@echo "$(YELLOW)💰 Computing Services:$(NC)"
	@echo "  start-free-service    - Start free PI service"
	@echo "  start-payment-service - Start payment service"
	@echo "  test-pi              - Test PI calculation"
	@echo "  benchmark-pi         - Run PI benchmark"
	@echo "  test-free-service    - Test free service API"
	@echo ""
	@echo "$(YELLOW)📊 Client Commands:$(NC)"
	@echo "  status         - Check client status"
	@echo "  whoami         - Show client identity"
	@echo "  balance        - Check account balance"
	@echo "  check-gpu      - Check GPU availability"
	@echo ""
	@echo "$(YELLOW)🔬 Scientific Features:$(NC)"
	@echo "  example-orbital      - Run orbital dynamics example"
	@echo "  example-gpu-training - Run GPU training example"
	@echo ""
	@echo "$(YELLOW)🐳 Docker & Release:$(NC)"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-build-gpu - Build Docker image with GPU"
	@echo "  release          - Create release build"
	@echo ""
	@echo "$(YELLOW)ℹ️  Information:$(NC)"
	@echo "  info           - Show build information"
	@echo "  help           - Show this help"
	@echo ""
	@echo "$(CYAN)Quick Start:$(NC)"
	@echo "  make dev                    # Full setup"
	@echo "  make start-free-service     # Start computing service"
	@echo "  make test-pi                # Test PI calculation"
