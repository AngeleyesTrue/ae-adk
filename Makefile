# AE-ADK Go Edition
# Build and development automation
# Note: On Windows, run via Git Bash or use 'go build/test/install' directly in PowerShell.

BINARY_NAME := ae
MODULE := github.com/AngeleyesTrue/ae-adk
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell go run -mod=readonly ./internal/cmd/datestamp 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-s -w -X $(MODULE)/pkg/version.Version=$(VERSION) -X $(MODULE)/pkg/version.Commit=$(COMMIT) -X $(MODULE)/pkg/version.Date=$(DATE)"

# Local release configuration
LOCAL_RELEASE_DIR ?= $(HOME)/.ae/releases
PLATFORM := $(shell go env GOOS)-$(shell go env GOARCH)
RELEASE_BINARY := ae-$(VERSION)-$(PLATFORM)

.PHONY: all build test lint fix clean install generate help release-local

all: lint test build ## Run lint, test, and build

build: ## Build the binary
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/ae

release-local: build ## Create a local release for development updates
	@echo "Creating local release: $(VERSION)"
	@mkdir -p $(LOCAL_RELEASE_DIR)
	@cp bin/$(BINARY_NAME) $(LOCAL_RELEASE_DIR)/$(RELEASE_BINARY)
	@chmod +x $(LOCAL_RELEASE_DIR)/$(RELEASE_BINARY) 2>/dev/null || true
	@echo '{"version":"$(VERSION)","date":"$(DATE)","platform":"$(PLATFORM)","binary":"$(RELEASE_BINARY)"}' > $(LOCAL_RELEASE_DIR)/version.json
	@echo "Local release created at: $(LOCAL_RELEASE_DIR)"
	@echo "  Binary: $(RELEASE_BINARY)"
	@echo "  Version: $(VERSION)"

install: ## Install the binary
	go install $(LDFLAGS) ./cmd/ae

test: ## Run tests with race detection
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

test-verbose: ## Run tests with verbose output
	go test -race -v -coverprofile=coverage.out -covermode=atomic ./...

coverage: test ## Show test coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters
	golangci-lint run ./...

fix: ## Run go fix modernizers (twice for synergistic fixes)
	go fix ./...
	go fix ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format code
	gofumpt -l -w .

generate: ## Run go generate
	go generate ./...

clean: ## Remove build artifacts
	go clean
	rm -rf bin/ coverage.out coverage.html 2>/dev/null || true

tidy: ## Tidy go modules
	go mod tidy

run: build ## Build and run
	./bin/$(BINARY_NAME)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
