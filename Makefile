.PHONY: build clean install test run help deps

BINARY_NAME=osc
GO=go
GOFLAGS=-v
INSTALL_PATH=/usr/local/bin

all: clean lint build ## Clean, lint, and build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	$(GO) mod download
	$(GO) mod tidy

build: deps ## Build the binary
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .

clean: ## Remove build artifacts
	$(GO) clean
	rm -f $(BINARY_NAME)

install: build ## Install binary to system
	install -m 755 $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

uninstall: ## Uninstall binary from system
	rm -f $(INSTALL_PATH)/$(BINARY_NAME)

test: ## Run tests
	$(GO) test -v ./...

run: build ## Build and run with health command
	./$(BINARY_NAME) health

fmt: ## Format code
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

lint: fmt vet ## Run formatters and linters

