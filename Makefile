# Makefile for httpbin

# Variables
BINARY_NAME=httpbin
OUTPUT_DIR=bin
MAIN_PATH=./cmd/httpbin
GO=/usr/local/go/bin/go
LDFLAGS=-s -w

# Build info
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

.PHONY: all build build-static build test lint fmt vet clean

all: build

build-static:
	@echo "Building static binary (CGO_ENABLED=0)..."
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 $(GO) build \
		-a \
		-ldflags "$(LDFLAGS) -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)-static \
		$(MAIN_PATH)
	@echo "Static binary built: $(OUTPUT_DIR)/$(BINARY_NAME)-static"
	@ls -lh $(OUTPUT_DIR)/$(BINARY_NAME)

build:
	@echo "Building with CGO_ENABLED=1..."
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=1 $(GO) build \
		-ldflags "$(LDFLAGS) -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME) \
		$(MAIN_PATH)
	@echo "CGO binary built: $(OUTPUT_DIR)/$(BINARY_NAME)"
	@ls -lh $(OUTPUT_DIR)/$(BINARY_NAME)


test:
	@echo "Running tests..."
	$(GO) test -v -race -cover ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: fmt vet
	@echo "Linting complete"

fmt:
	@echo "Checking code formatting..."
	$(GO) fmt -n ./...

vet:
	@echo "Running go vet..."
	$(GO) vet ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(OUTPUT_DIR)
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	$(GO) clean
	@echo "Clean complete"

clean-cache:
	@echo "Cleaning all build caches..."
	$(GO) clean -cache -testcache

run: build
	@$(OUTPUT_DIR)/$(BINARY_NAME)
