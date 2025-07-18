# Makefile for Redi Frontend Server

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := redi
REJS_BINARY := rejs
BUILD_BINARY := redi-build
BUILD_DIR := .
FIXTURES_DIR := fixtures
PORT := 8080

# Version information
VERSION := $(shell git describe --tags --exact-match 2>/dev/null || git describe --tags --always 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS := -X main.Version=$(VERSION)
REJS_LDFLAGS := -X main.Version=$(VERSION)
BUILD_LDFLAGS := -X main.Version=$(VERSION)

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
	@echo "\033[32mbuild-all\033[0m      Build all binaries (redi, rejs, redi-build)"
	@echo "\033[32mrun\033[0m            Run the server directly with test fixtures"
	@echo "\033[32mstart\033[0m          Build and run the server with test fixtures"
	@echo "\033[32mtest\033[0m           Run all tests"
	@echo "\033[32mtest-unit\033[0m      Run unit tests only"
	@echo "\033[32mtest-integration\033[0m Run integration tests only"
	@echo "\033[32mtest-api\033[0m       Run API tests only"
	@echo "\033[32mtest-e2e\033[0m       Run E2E tests with Puppeteer"
	@echo "\033[32mtest-e2e-debug\033[0m Run E2E tests with visible browser"
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

## build-rejs: Build the rejs JavaScript runtime binary
.PHONY: build-rejs
build-rejs:
	@echo "$(YELLOW)Building $(REJS_BINARY) version $(VERSION)...$(RESET)"
	@go build -ldflags="$(REJS_LDFLAGS)" -o $(BUILD_DIR)/$(REJS_BINARY) ./cmd/rejs
	@echo "$(GREEN)✅ Build completed: $(BUILD_DIR)/$(REJS_BINARY)$(RESET)"

## build-redi-build: Build the redi-build tool binary
.PHONY: build-redi-build
build-redi-build:
	@echo "$(YELLOW)Building $(BUILD_BINARY) version $(VERSION)...$(RESET)"
	@go build -ldflags="$(BUILD_LDFLAGS)" -o $(BUILD_DIR)/$(BUILD_BINARY) ./cmd/redi-build
	@echo "$(GREEN)✅ Build completed: $(BUILD_DIR)/$(BUILD_BINARY)$(RESET)"

## build-all: Build all binaries (redi, rejs, redi-build)
.PHONY: build-all
build-all: build build-rejs build-redi-build
	@echo "$(GREEN)✅ All binaries built$(RESET)"

## clean: Remove built binaries and temporary files
.PHONY: clean
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	@rm -f $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/$(REJS_BINARY) $(BUILD_DIR)/$(BUILD_BINARY)
	@go clean
	@echo "$(GREEN)✅ Clean completed$(RESET)"

## run: Run the server directly with test fixtures (without building binary)
.PHONY: run
run:
	@echo "$(YELLOW)Stopping any existing server on port $(PORT)...$(RESET)"
	@lsof -ti:$(PORT) | xargs kill -9 2>/dev/null || true
	@sleep 1
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
	@go test -v -count=1 ./...
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

## test-e2e: Run E2E tests with Puppeteer
.PHONY: test-e2e
test-e2e: build
	@echo "$(YELLOW)Running E2E tests with Puppeteer...$(RESET)"
	@if command -v puppeteer >/dev/null 2>&1 || npm list puppeteer >/dev/null 2>&1; then \
		node e2e/run-tests.js; \
	else \
		echo "$(RED)❌ Puppeteer not found. Install with: npm install -g puppeteer$(RESET)"; \
		exit 1; \
	fi

## test-e2e-watch: Run E2E tests in watch mode
.PHONY: test-e2e-watch
test-e2e-watch: build
	@echo "$(YELLOW)Running E2E tests in watch mode...$(RESET)"
	@node e2e/run-tests.js --watch

## test-e2e-debug: Run E2E tests with visible browser
.PHONY: test-e2e-debug
test-e2e-debug: build
	@echo "$(YELLOW)Running E2E tests with visible browser...$(RESET)"
	@PUPPETEER_HEADLESS=false node e2e/run-tests.js

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

## install: Install all binaries to GOPATH/bin
.PHONY: install
install:
	@echo "$(YELLOW)Installing all binaries version $(VERSION)...$(RESET)"
	@go install -ldflags="$(LDFLAGS)" ./cmd/redi
	@go install -ldflags="$(REJS_LDFLAGS)" ./cmd/rejs
	@go install -ldflags="$(BUILD_LDFLAGS)" ./cmd/redi-build
	@echo "$(GREEN)✅ All binaries installed to GOPATH/bin$(RESET)"

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

## release: Build release version with optimizations for all binaries
.PHONY: release
release: clean
	@echo "$(YELLOW)Building release version $(VERSION)...$(RESET)"
	@CGO_ENABLED=0 go build -ldflags="-w -s $(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/redi
	@CGO_ENABLED=0 go build -ldflags="-w -s $(REJS_LDFLAGS)" -o $(BUILD_DIR)/$(REJS_BINARY) ./cmd/rejs
	@CGO_ENABLED=0 go build -ldflags="-w -s $(BUILD_LDFLAGS)" -o $(BUILD_DIR)/$(BUILD_BINARY) ./cmd/redi-build
	@echo "$(GREEN)✅ Release build completed for all binaries$(RESET)"

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
	@echo "$(BLUE)Redi Toolkit version $(VERSION)$(RESET)"
	@echo "Build date: $(BUILD_DATE)"
	@echo "Git commit: $(GIT_COMMIT)"
	@echo ""
	@echo "$(YELLOW)Available binaries:$(RESET)"
	@echo "  - redi (Web Server)"
	@echo "  - rejs (JavaScript Runtime)"
	@echo "  - redi-build (Build Tools)"

.PHONY: all start
all: clean fmt vet test build-all