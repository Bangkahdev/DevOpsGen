# DevOpsGen build and test automation.
#
# Targets follow the SRS roadmap phases:
#   build   - compile the binary for the current platform
#   test    - run unit and integration tests
#   lint    - static analysis
#   tidy    - synchronize go.mod/go.sum

BINARY := devopsgen
CMD := ./cmd/devopsgen
GO := go
GOFLAGS := -mod=mod

.PHONY: build test test-coverage lint tidy clean help

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make [target]\n\nTargets:\n"} \
	/^[a-zA-Z_-]+:.*?##/ { printf "  %-15s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: tidy ## Build the binary for the current platform
	$(GO) build $(GOFLAGS) -o $(BINARY) $(CMD)

test: tidy ## Run all tests
	$(GO) test ./...

test-coverage: tidy ## Run tests with coverage report
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

lint: tidy ## Run static analysis (go vet + staticcheck when available)
	$(GO) vet ./...
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed; skipping"

tidy: ## Sync go.mod and go.sum
	$(GO) mod tidy

clean: ## Remove build artifacts
	rm -f $(BINARY) coverage.out