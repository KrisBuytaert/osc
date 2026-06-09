.PHONY: build clean install test run help deps

BINARY_NAME=osc
GO=go
GOFLAGS=-v
INSTALL_PATH=/usr/local/bin
LDFLAGS=-s -w -extldflags '-static'
BUILD_TAGS=netgo osusergo

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
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) \
		-tags '$(BUILD_TAGS)' \
		-ldflags="$(LDFLAGS)" \
		-a \
		-o $(BINARY_NAME) .


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

COMPOSE_FILE    := tests/integration/docker-compose.yml
COMPOSE_PROJECT := osc-inttest
BATS_SUITE      := tests/integration

.PHONY: test-integration test-integration-teardown

test-integration: build ## Run integration tests against a local 2-node OpenSearch cluster
	@command -v bats >/dev/null 2>&1 || \
	  { echo "bats not found. Install: https://github.com/bats-core/bats-core"; exit 1; }
	@command -v docker >/dev/null 2>&1 || \
	  { echo "docker not found"; exit 1; }
	@echo "==> Starting OpenSearch cluster..."
	docker compose -p $(COMPOSE_PROJECT) -f $(COMPOSE_FILE) up -d --wait
	@echo "==> Waiting for cluster health (yellow or better)..."
	@until curl -sf "http://localhost:19200/_cluster/health?wait_for_status=yellow&timeout=5s" >/dev/null 2>&1; do \
		sleep 3; echo "    still waiting..."; \
	done
	@echo "==> Running bats suite..."
	OSC="$(PWD)/$(BINARY_NAME)" bats --tap $(BATS_SUITE); \
	EXIT_CODE=$$?; \
	echo "==> Tearing down..."; \
	docker compose -p $(COMPOSE_PROJECT) -f $(COMPOSE_FILE) down -v; \
	exit $$EXIT_CODE

test-integration-teardown: ## Force remove integration test containers and volumes
	docker compose -p $(COMPOSE_PROJECT) -f $(COMPOSE_FILE) down -v --remove-orphans

