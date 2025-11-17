.PHONY: help build build-api build-cli build-tui run run-tui test clean docker-build docker-clean docker-up docker-down docker-logs docker-restart docker-ps deps install fmt lint lint-fix lint-verbose test-unit test-integration test-coverage docker-test all

# Variables
API_BINARY=will-it-compile-api
CLI_BINARY=will-it-compile
TUI_BINARY=will-it-compile-tui
GO=go
GOFLAGS=-v
DOCKER=docker
DOCKER_COMPOSE=docker compose
GOLANGCI_LINT=golangci-lint

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
	MINIMAL_IMAGE_VALIDATION=true $(GO) test -v ./tests/integration/

test: docker-build ## Run all tests (builds Docker images first)
	$(GO) test -v ./...

test-coverage: docker-build ## Run tests with coverage
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
	$(GO) clean

docker-pull: ## Pull official compiler images for local testing
	@echo "Pulling official compiler images..."
	@echo "→ GCC (C/C++ - Debian-based)..."
	@$(DOCKER) pull gcc:9
	@$(DOCKER) pull gcc:11
	@$(DOCKER) pull gcc:13
	@echo "→ Go (Alpine-based)..."
	@$(DOCKER) pull golang:1.22-alpine
	@$(DOCKER) pull golang:1.23-alpine
	@echo "→ Rust (Alpine-based)..."
	@$(DOCKER) pull rust:1.75-alpine
	@$(DOCKER) pull rust:1.80-alpine
	@echo "✓ All compiler images pulled"

docker-build: docker-pull ## Pull compiler images (alias for backward compatibility)

docker-clean: ## Remove Docker images
	@echo "Removing official compiler images..."
	@$(DOCKER) rmi gcc:9 gcc:11 gcc:13 || true
	@$(DOCKER) rmi golang:1.22-alpine golang:1.23-alpine || true
	@$(DOCKER) rmi rust:1.75-alpine rust:1.80-alpine || true
	@echo "✓ Cleanup complete"

docker-test: docker-pull ## Test Docker image with official GCC
	@echo "Testing official GCC Docker image..."
	@echo '#include <iostream>\nint main() { std::cout << "Hello from official GCC!" << std::endl; return 0; }' > /tmp/test.cpp
	@$(DOCKER) run --rm \
		-v /tmp/test.cpp:/workspace/source.cpp:ro \
		-w /workspace \
		gcc:13 \
		sh -c 'g++ -std=c++17 source.cpp -o output && ./output'
	@rm /tmp/test.cpp
	@echo "✓ Docker test passed"

docker-up: ## Start docker compose services
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop docker compose services and remove volumes
	$(DOCKER_COMPOSE) down -v

docker-logs: ## View docker compose logs
	$(DOCKER_COMPOSE) logs -f

docker-restart: ## Restart docker compose services
	$(DOCKER_COMPOSE) restart

docker-ps: ## Show running docker compose containers
	$(DOCKER_COMPOSE) ps

fmt: ## Format Go code
	$(GO) fmt ./...

lint: ## Run golangci-lint
	@which $(GOLANGCI_LINT) > /dev/null || (echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	$(GOLANGCI_LINT) run ./...

lint-fix: ## Run golangci-lint with auto-fix
	@which $(GOLANGCI_LINT) > /dev/null || (echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	$(GOLANGCI_LINT) run --fix ./...

lint-verbose: ## Run golangci-lint with verbose output
	@which $(GOLANGCI_LINT) > /dev/null || (echo "golangci-lint not installed. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	$(GOLANGCI_LINT) run -v ./...

all: clean deps docker-build build test ## Run all build steps
