# Makefile for Redi Frontend Server

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := redi
BUILD_DIR := .
FIXTURES_DIR := fixtures
PORT := 8080

# Version information
VERSION := $(shell git describe --tags --exact-match 2>/dev/null || git describe --tags --always 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -X main.Version=$(VERSION)

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
BLUE := \033[34m
RESET := \033[0m

## help: Show this help message
.PHONY: help
help:
	@echo "\033[34mRedi Frontend Server - Available Commands\033[0m"
	@echo "========================================"
	@echo "\033[32mbuild\033[0m          Build the redi binary"
	@echo "\033[32mrun\033[0m            Run the server directly with test fixtures"
	@echo "\033[32mstart\033[0m          Build and run the server with test fixtures"
	@echo "\033[32mtest\033[0m           Run all tests"
	@echo "\033[32mtest-unit\033[0m      Run unit tests only"
	@echo "\033[32mtest-integration\033[0m Run integration tests only"
	@echo "\033[32mtest-api\033[0m       Run API tests only"
	@echo "\033[32mbench\033[0m          Run benchmark tests"
	@echo "\033[32mcoverage\033[0m       Run tests with coverage report"
	@echo "\033[32mfmt\033[0m            Format Go code"
	@echo "\033[32mvet\033[0m            Run go vet"
	@echo "\033[32mlint\033[0m           Run golangci-lint"
	@echo "\033[32mdeps\033[0m           Download and tidy dependencies"
	@echo "\033[32minstall\033[0m        Install the binary to GOPATH/bin"
	@echo "\033[32mdev\033[0m            Development mode with file watching"
	@echo "\033[32mcheck\033[0m          Run all checks (fmt, vet, test)"
	@echo "\033[32mrelease\033[0m        Build release version with optimizations"
	@echo "\033[32mclean\033[0m          Remove built binaries and temporary files"
	@echo "\033[32mfixtures-list\033[0m  List all fixtures and routes"
	@echo "\033[32minit\033[0m           Initialize project dependencies and tools"
	@echo "\033[32mversion\033[0m        Show version information"
	@echo ""

## build: Build the redi binary
.PHONY: build
build:
	@echo "$(YELLOW)Building $(BINARY_NAME) version $(VERSION)...$(RESET)"
	@go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/redi
	@echo "$(GREEN)✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

## clean: Remove built binaries and temporary files
.PHONY: clean
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@go clean
	@echo "$(GREEN)✅ Clean completed$(RESET)"

## run: Run the server directly with test fixtures (without building binary)
.PHONY: run
run:
	@echo "$(YELLOW)Starting redi server...$(RESET)"
	@echo "$(BLUE)Server available at: http://localhost:$(PORT)$(RESET)"
	@echo "$(BLUE)Press Ctrl+C to stop$(RESET)"
	@go run ./cmd/redi --root=$(FIXTURES_DIR) --port=$(PORT)

## start: Build and run the server with test fixtures
.PHONY: start
start: build
	@echo "$(YELLOW)Starting redi server (built binary)...$(RESET)"
	@echo "$(BLUE)Server available at: http://localhost:$(PORT)$(RESET)"
	@echo "$(BLUE)Press Ctrl+C to stop$(RESET)"
	@./$(BINARY_NAME) --root=$(FIXTURES_DIR) --port=$(PORT)

## test: Run all tests
.PHONY: test
test:
	@echo "$(YELLOW)Running all tests...$(RESET)"
	@go test -v ./...
	@echo "$(GREEN)✅ All tests completed$(RESET)"

## test-unit: Run unit tests only
.PHONY: test-unit
test-unit:
	@echo "$(YELLOW)Running unit tests...$(RESET)"
	@go test -v -run "Test(Server|Route|Static|Markdown|JavaScript|HTML|Dynamic|Layout|Script)" ./...

## test-integration: Run integration tests only
.PHONY: test-integration
test-integration:
	@echo "$(YELLOW)Running integration tests...$(RESET)"
	@go test -v -run "Integration" ./...

## test-api: Run API tests only
.PHONY: test-api
test-api:
	@echo "$(YELLOW)Running API tests...$(RESET)"
	@go test -v -run "API" ./...

## bench: Run benchmark tests
.PHONY: bench
bench:
	@echo "$(YELLOW)Running benchmarks...$(RESET)"
	@go test -bench=. -benchmem ./...

## coverage: Run tests with coverage report
.PHONY: coverage
coverage:
	@echo "$(YELLOW)Running tests with coverage...$(RESET)"
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report generated: coverage.html$(RESET)"

## fmt: Format Go code
.PHONY: fmt
fmt:
	@echo "$(YELLOW)Formatting Go code...$(RESET)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(RESET)"

## vet: Run go vet
.PHONY: vet
vet:
	@echo "$(YELLOW)Running go vet...$(RESET)"
	@go vet ./...
	@echo "$(GREEN)✅ Go vet completed$(RESET)"

## lint: Run golangci-lint (requires golangci-lint to be installed)
.PHONY: lint
lint:
	@echo "$(YELLOW)Running golangci-lint...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "$(GREEN)✅ Linting completed$(RESET)"; \
	else \
		echo "$(RED)❌ golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(RESET)"; \
	fi

## deps: Download and tidy dependencies
.PHONY: deps
deps:
	@echo "$(YELLOW)Downloading dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✅ Dependencies updated$(RESET)"

## install: Install the binary to GOPATH/bin
.PHONY: install
install:
	@echo "$(YELLOW)Installing $(BINARY_NAME) version $(VERSION)...$(RESET)"
	@go install -ldflags="$(LDFLAGS)" ./cmd/redi
	@echo "$(GREEN)✅ $(BINARY_NAME) installed to GOPATH/bin$(RESET)"

## dev: Development mode - build and run with file watching (requires entr or similar)
.PHONY: dev
dev:
	@echo "$(YELLOW)Starting development mode...$(RESET)"
	@if command -v entr >/dev/null 2>&1; then \
		echo "$(BLUE)Watching for file changes... (Press Ctrl+C to stop)$(RESET)"; \
		find . -name "*.go" | entr -r make run; \
	else \
		echo "$(RED)❌ entr not found. Install with: brew install entr (macOS) or apt-get install entr (Ubuntu)$(RESET)"; \
		echo "$(YELLOW)Falling back to regular run...$(RESET)"; \
		make run; \
	fi

## check: Run all checks (fmt, vet, test)
.PHONY: check
check: fmt vet test
	@echo "$(GREEN)✅ All checks passed$(RESET)"

## release: Build release version with optimizations
.PHONY: release
release: clean
	@echo "$(YELLOW)Building release version $(VERSION)...$(RESET)"
	@CGO_ENABLED=0 go build -ldflags="-w -s $(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/redi
	@echo "$(GREEN)✅ Release build completed$(RESET)"

## docker-build: Build Docker image (requires Dockerfile)
.PHONY: docker-build
docker-build:
	@echo "$(YELLOW)Building Docker image...$(RESET)"
	@if [ -f Dockerfile ]; then \
		docker build -t redi:latest .; \
		echo "$(GREEN)✅ Docker image built: redi:latest$(RESET)"; \
	else \
		echo "$(RED)❌ Dockerfile not found$(RESET)"; \
	fi

## fixtures-list: List all fixtures and routes
.PHONY: fixtures-list
fixtures-list:
	@echo "$(BLUE)Available fixtures and routes:$(RESET)"
	@echo "$(YELLOW)Static files:$(RESET)"
	@find $(FIXTURES_DIR)/public -type f 2>/dev/null | sed 's|$(FIXTURES_DIR)/public|  http://localhost:$(PORT)|' || echo "  No static files found"
	@echo "$(YELLOW)Routes:$(RESET)"
	@find $(FIXTURES_DIR)/routes -name "*.html" -o -name "*.js" -o -name "*.md" 2>/dev/null | \
		grep -v "_layout" | \
		sed 's|$(FIXTURES_DIR)/routes||' | \
		sed 's|/index\.html$$|/|' | \
		sed 's|\.html$$||' | \
		sed 's|\.js$$||' | \
		sed 's|\.md$$||' | \
		sed 's|\[id\]|{id}|g' | \
		sed 's|^|  http://localhost:$(PORT)|' || echo "  No routes found"

## serve-docs: Serve documentation (if available)
.PHONY: serve-docs
serve-docs:
	@echo "$(YELLOW)Starting documentation server...$(RESET)"
	@if [ -f "coverage.html" ]; then \
		echo "$(BLUE)Coverage report: http://localhost:8081/coverage.html$(RESET)"; \
		python3 -m http.server 8081 2>/dev/null || python -m SimpleHTTPServer 8081; \
	else \
		echo "$(RED)❌ No documentation found. Run 'make coverage' first.$(RESET)"; \
	fi

## init: Initialize project dependencies and tools
.PHONY: init
init:
	@echo "$(YELLOW)Initializing project...$(RESET)"
	@go mod tidy
	@echo "$(GREEN)✅ Project initialized$(RESET)"
	@echo "$(BLUE)Recommended tools to install:$(RESET)"
	@echo "  - golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
	@echo "  - entr (for dev mode): brew install entr (macOS) or apt-get install entr (Ubuntu)"

# File dependencies
$(BINARY_NAME): cmd/redi/*.go *.go handlers/*.go modules/**/*.go
	@make build

## version: Show version information
.PHONY: version
version:
	@echo "$(BLUE)redi version $(VERSION)$(RESET)"
	@echo "Build date: $(BUILD_DATE)"
	@echo "Git commit: $(GIT_COMMIT)"

.PHONY: all start
all: clean fmt vet test build