# Makefile for httpbin

# Variables
BINARY_NAME=httpbin
OUTPUT_DIR=bin
MAIN_PATH=./cmd/httpbin
GO?=go
LDFLAGS=-s -w

# Build info
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

.PHONY: all build build-static build-fips build test lint fmt vet clean ldflags snapshot build-build-image smoke-test-rpm helm-lint

all: build build-static build-fips

ldflags:
	@echo "$(LDFLAGS) -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT)"

LDFLAGS_FULL := $(LDFLAGS) -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT)

build-static:
	@echo "Building static binary (CGO_ENABLED=0)..."
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 $(GO) build \
		-a \
		-ldflags "$(LDFLAGS) -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)-static \
		$(MAIN_PATH)
	@echo "Static binary built: $(OUTPUT_DIR)/$(BINARY_NAME)-static"
	@ls -lh $(OUTPUT_DIR)/$(BINARY_NAME)-static

build:
	@echo "Building with CGO_ENABLED=1..."
	@mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=1 $(GO) build \
		-ldflags "$(LDFLAGS) -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME) \
		$(MAIN_PATH)
	@echo "CGO binary built: $(OUTPUT_DIR)/$(BINARY_NAME)"
	@ls -lh $(OUTPUT_DIR)/$(BINARY_NAME)

build-fips:
	@echo "Building FIPS-compliant binary (GOEXPERIMENT=boringcrypto)..."
	@mkdir -p $(OUTPUT_DIR)
	GOEXPERIMENT=boringcrypto CGO_ENABLED=1 $(GO) build \
		-tags fips \
		-ldflags "$(LDFLAGS) -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT)" \
		-o $(OUTPUT_DIR)/$(BINARY_NAME)-fips \
		$(MAIN_PATH)
	@echo "FIPS binary built: $(OUTPUT_DIR)/$(BINARY_NAME)-fips"
	@ls -lh $(OUTPUT_DIR)/$(BINARY_NAME)-fips


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

build-build-image:
	@echo "Building RHEL 8 build image..."
	docker build -f ci/Dockerfile.rhel8-build -t rhel8-goreleaser .

snapshot: build-build-image
	@echo "Creating GoReleaser snapshot inside RHEL 8 container..."
	docker run --rm \
		-v "$(shell pwd):/src" \
		-w /src \
		-e GITHUB_REPOSITORY=olamilekan000/httpbin-ops \
		-e LDFLAGS="$(LDFLAGS_FULL)" \
		rhel8-goreleaser \
		goreleaser release --snapshot --clean --skip=sign

smoke-test-rpm:
	@echo "Smoke-testing RPMs in Docker containers..."
	@if [ ! -d "dist" ]; then echo "Error: dist/ directory not found. Run 'make snapshot' first."; exit 1; fi
	@for arch in amd64 arm64; do \
		pattern="httpbin-dynamic*_linux_$$arch.rpm"; \
		rpm_file=$$(find dist -name "$$pattern" | head -n 1); \
		if [ -z "$$rpm_file" ]; then \
			echo "Skipping $$arch (no RPM found)"; \
			continue; \
		fi; \
		img_arch="$$arch"; [ "$$arch" = "amd64" ] && img_arch="x86_64"; \
		[ "$$arch" = "arm64" ] && img_arch="aarch64"; \
		for image in rockylinux:8 rockylinux:9 fedora:latest; do \
			echo "--- Testing $$arch RPM on $$image ---"; \
			docker run --rm --platform linux/$$img_arch \
				-v $$(pwd)/dist:/dist:ro \
				$$image \
				bash -c "rpm -ivh --nosignature /dist/$${rpm_file#dist/} && httpbin -version"; \
		done; \
	done

helm-lint:
	@echo "Linting Helm chart..."
	helm lint httpbin-ops-charts
