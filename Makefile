.PHONY: help build build-api build-cli build-tui run run-tui test clean docker-build docker-clean docker-up docker-down docker-logs docker-restart docker-ps deps install

# Variables
API_BINARY=will-it-compile-api
CLI_BINARY=will-it-compile
TUI_BINARY=will-it-compile-tui
GO=go
GOFLAGS=-v
DOCKER=docker

# Version information (can be overridden via build flags)
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X github.com/stlpine/will-it-compile/cmd/cli/commands.version=$(VERSION) \
                   -X github.com/stlpine/will-it-compile/cmd/cli/commands.commit=$(COMMIT) \
                   -X github.com/stlpine/will-it-compile/cmd/cli/commands.buildDate=$(BUILD_DATE)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download Go dependencies
	$(GO) mod download
	$(GO) mod verify

build-api: deps ## Build the API server only
	$(GO) build $(GOFLAGS) -o bin/$(API_BINARY) cmd/api/main.go

build-cli: deps ## Build the CLI tool only
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(CLI_BINARY) cmd/cli/main.go

build-tui: deps ## Build the TUI client only
	$(GO) build $(GOFLAGS) -o bin/$(TUI_BINARY) cmd/tui/main.go

build: build-api build-cli build-tui ## Build API server, CLI tool, and TUI client

run: build-api ## Build and run the API server
	./bin/$(API_BINARY)

run-tui: build-tui ## Build and run the TUI client
	./bin/$(TUI_BINARY)

install: build-cli ## Install CLI to $GOPATH/bin
	cp bin/$(CLI_BINARY) $(GOPATH)/bin/$(CLI_BINARY)
	@echo "✓ Installed $(CLI_BINARY) to $(GOPATH)/bin/"

test-unit: ## Run unit tests only (no Docker required)
	$(GO) test -v -short ./...

test-integration: docker-build ## Run integration tests (requires Docker)
	$(GO) test -v ./tests/integration/

test: docker-build ## Run all tests (builds Docker images first)
	$(GO) test -v ./...

test-coverage: docker-build ## Run tests with coverage
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	$(GO) clean

docker-build: ## Build Docker image for C++ compilation
	@echo "Building C++ compiler Docker image..."
	cd images/cpp && chmod +x build.sh && ./build.sh

docker-clean: ## Remove Docker images
	$(DOCKER) rmi will-it-compile/cpp-gcc:13-alpine || true

docker-test: docker-build ## Test Docker image
	@echo "Testing Docker image..."
	@echo '#include <iostream>\nint main() { std::cout << "Hello, World!" << std::endl; return 0; }' > /tmp/test.cpp
	@$(DOCKER) run --rm \
		-v /tmp/test.cpp:/workspace/source.cpp:ro \
		will-it-compile/cpp-gcc:13-alpine
	@rm /tmp/test.cpp

docker-up: docker-build ## Start docker compose services (builds compiler image first)
	docker compose up -d

docker-down: ## Stop docker compose services
	docker compose down

docker-logs: ## View docker compose logs
	docker compose logs -f

docker-restart: ## Restart docker compose services
	docker compose restart

docker-ps: ## Show running docker compose containers
	docker compose ps

fmt: ## Format Go code
	$(GO) fmt ./...

lint: ## Run golangci-lint
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

lint-fix: ## Run golangci-lint with auto-fix
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run --fix ./...

lint-verbose: ## Run golangci-lint with verbose output
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run -v ./...

dev: docker-build ## Set up development environment
	@echo "Development environment ready!"
	@echo "Run 'make run' to start the server"

all: clean deps docker-build build test ## Run all build steps
